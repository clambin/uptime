package main

import (
	"errors"
	"flag"
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
	debug    = flag.Bool("debug", false, "Log debugging information")
	token    = flag.String("token", "", "Authorization token")
	addr     = flag.String("addr", ":8080", "Listener port")
	promAddr = flag.String("prom", ":9090", "Prometheus metrics port")
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

	metrics := monitor.NewHTTPMetrics("uptime", "monitor")
	prometheus.MustRegister(metrics)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(*promAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	s := http.Server{
		Addr: *addr,
		Handler: auth.Authenticate(*token)(
			logger.WithLogger(l)(
				monitor.New(metrics, nil),
			),
		),
	}

	if err := s.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
