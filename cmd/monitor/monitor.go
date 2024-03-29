package main

import (
	"errors"
	"flag"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/uptime/internal/monitor"
	monitorMetrics "github.com/clambin/uptime/internal/monitor/metrics"
	"github.com/clambin/uptime/pkg/auth"
	"github.com/clambin/uptime/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"os"
)

var (
	version  = "change-me"
	debug    = flag.Bool("debug", false, "Log debugging information")
	token    = flag.String("token", "", "Authorization token")
	addr     = flag.String("addr", ":8080", "Listener port")
	promAddr = flag.String("prom", ":9090", "Prometheus metrics port")

	clientMetricBuckets = prometheus.DefBuckets
)

func main() {
	flag.Parse()

	var opts slog.HandlerOptions
	if *debug {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewJSONHandler(os.Stdout, &opts))

	if *token == "" {
		l.Warn("no token provided")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(*promAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	serverMetrics := metrics.NewRequestSummaryMetrics("uptime", "monitor_server", nil)
	monMetrics := monitorMetrics.NewHostMetrics("uptime", "monitor_target", nil)
	httpClientMetrics := monitorMetrics.NewHTTPMetrics("uptime", "monitor_target", nil, clientMetricBuckets...)
	prometheus.MustRegister(httpClientMetrics, serverMetrics, monMetrics)

	h := monitor.New(
		monMetrics,
		&http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: roundtripper.New(roundtripper.WithRequestMetrics(httpClientMetrics)),
			Timeout:   monitor.DefaultClientTimeout,
		},
	)

	if *token != "" {
		h = auth.Authenticate(*token)(h)
	}

	s := http.Server{
		Addr: *addr,
		Handler: middleware.WithRequestMetrics(serverMetrics)(
			logger.WithLogger(l)(
				h,
			),
		),
	}

	l.Info("starting uptime monitor", "version", version)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
	l.Info("uptime monitor stopped")
}
