package main

import (
	"context"
	"errors"
	"flag"
	"github.com/clambin/uptime/internal/agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"syscall"
)

var (
	version       = "change-me"
	debug         = flag.Bool("debug", false, "log debug messages")
	monitor       = flag.String("monitor", "", "host monitor URL (required)")
	token         = flag.String("token", "", "host monitor token (required)")
	promAddr      = flag.String("prom", ":9090", "Prometheus metrics port")
	configuration = flag.String("configuration", "", "configuration file")
)

func main() {
	flag.Parse()

	var cfg agent.Configuration
	if *configuration != "" {
		var err error
		if cfg, err = agent.LoadFromFile(*configuration); err != nil {
			panic(err)
		}
	}
	if *monitor != "" {
		cfg.Monitor = *monitor
	}
	if *token != "" {
		cfg.Token = *token
	}

	var opts slog.HandlerOptions
	if *debug {
		opts.Level = slog.LevelDebug
	}
	l := slog.New(slog.NewJSONHandler(os.Stderr, &opts))

	c, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		l.Error("failed to connect to cluster", "err", err)
		return
	}

	metrics := agent.NewMetrics()
	prometheus.MustRegister(metrics)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(*promAddr, nil); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	a, err := agent.New(c, cfg, metrics, l)
	if err != nil {
		l.Error("failed to start agent", "err", err)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	l.Info("starting uptime agent", "version", version)
	a.Run(ctx)
	l.Info("uptime agent stopped")
}
