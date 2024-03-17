package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/uptime/internal/agent/informer"
	v1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"log/slog"
	"net/http"
	"time"
)

type Agent struct {
	ingressInformer *informer.Informer
	filter          filter
	reSender        reSender
	sender          sender
}

func New(c *kubernetes.Clientset, cfg Configuration, metrics *Metrics, logger *slog.Logger) (*Agent, error) {
	g := cache.NewListWatchFromClient(c.NetworkingV1().RESTClient(), "ingresses", v1.NamespaceAll, fields.Everything())
	return NewWithListWatcher(g, cfg, metrics, logger)
}

const (
	resyncPeriod = 5 * time.Minute
)

func NewWithListWatcher(lw cache.ListerWatcher, cfg Configuration, metrics *Metrics, logger *slog.Logger) (*Agent, error) {
	if cfg.Monitor == "" {
		return nil, errors.New("missing monitor URL")
	}

	filterIn := make(chan event)
	reSenderIn := make(chan event)
	senderIn := make(chan event)

	i, err := informer.New(lw, resyncPeriod, new(netv1.Ingress), &ingressWatcher{
		out:     filterIn,
		metrics: metrics,
		logger:  logger.With("component", "informer"),
	})
	if err != nil {
		return nil, fmt.Errorf("informer: %w", err)
	}

	return &Agent{
		ingressInformer: i,
		filter: filter{
			in:            filterIn,
			out:           reSenderIn,
			configuration: cfg,
			logger:        logger.With("component", "filter"),
		},
		reSender: reSender{
			in:     reSenderIn,
			out:    senderIn,
			events: make(map[string]event),
		},
		sender: sender{
			in:            senderIn,
			configuration: cfg,
			metrics:       metrics,
			httpClient:    http.DefaultClient,
			logger:        logger.With("component", "sender"),
		},
	}, nil
}

const senderCount = 5
const reSendInterval = 5 * time.Minute

func (a *Agent) Run(ctx context.Context) {
	for range senderCount {
		go a.sender.Run(ctx)
	}
	go a.reSender.Run(ctx, reSendInterval)
	go a.filter.Run(ctx)
	go a.ingressInformer.Run()
	defer a.ingressInformer.Cancel()
	<-ctx.Done()
}
