package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/clambin/uptime/internal/agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
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

	c, err := kubernetes.NewForConfig(getConfigOrDie(l))
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

func getConfigOrDie(l *slog.Logger) *rest.Config {
	cfg, err := getConfig()
	if err != nil {
		l.Error("unable to get kubeconfig", "err", err)
		os.Exit(1)
	}
	return cfg
}

func getConfig() (*rest.Config, error) {
	// try in-cluster config first
	c, err := rest.InClusterConfig()
	if err == nil {
		return c, nil
	}

	// try to get the user config
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if _, ok := os.LookupEnv("HOME"); !ok {
		u, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %w", err)
		}
		loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}
