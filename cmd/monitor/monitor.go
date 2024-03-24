package main

import (
	"errors"
	"flag"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/uptime/internal/monitor"
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

	clientMetricBuckets = []float64{0.25, 0.5, 0.75, 1, 2, 5, 10}
)

func main() {
	flag.Parse()

	var opts slog.HandlerOptions
	if *debug {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewJSONHandler(os.Stdout, &opts))

	if *token == "" {
		l.Error("no token provided")
		return
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(*promAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	httpClientMetrics := metrics.NewRequestHistogramMetrics("uptime", "monitor", map[string]string{"component": "client"}, clientMetricBuckets...)
	serverMetrics := metrics.NewRequestSummaryMetrics("uptime", "monitor", map[string]string{"component": "server"})
	monitorMetrics := monitor.NewHTTPMetrics()
	prometheus.MustRegister(httpClientMetrics, serverMetrics, monitorMetrics)

	s := http.Server{
		Addr: *addr,
		Handler: auth.Authenticate(*token)(
			middleware.WithRequestMetrics(serverMetrics)(
				logger.WithLogger(l)(
					monitor.New(monitorMetrics, &http.Client{
						Transport: roundtripper.New(roundtripper.WithRequestMetrics(httpClientMetrics)),
						Timeout:   monitor.DefaultClientTimeout,
					}),
				),
			),
		),
	}
	l.Info("starting uptime monitor", "version", version)
	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
	l.Info("uptime monitor stopped")
}
