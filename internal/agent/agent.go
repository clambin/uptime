package agent

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/uptime/internal/agent/informer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http"
	"time"
)

type Event struct {
	Type        EventType
	Host        string
	Annotations map[string]string
}

type EventType string

const (
	AddEvent    EventType = "ADD"
	DeleteEvent EventType = "DELETE"
)

type Agent struct {
	ingressInformer *informer.Informer
	filter          filter
	resender        reSender
	sender          sender
}

func New(c *kubernetes.Clientset, cfg Configuration, logger *slog.Logger) (*Agent, error) {
	if cfg.Monitor == "" {
		return nil, errors.New("missing monitor URL")
	}
	filterIn := make(chan Event)
	reSenderIn := make(chan Event)
	senderIn := make(chan Event)
	i, err := informer.NewIngressInformer(c, v1.NamespaceAll, &ingressWatcher{
		out:    filterIn,
		logger: logger.With("component", "watcher"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes informer: %w", err)
	}
	return &Agent{
		ingressInformer: i,
		filter: filter{
			in:            filterIn,
			out:           reSenderIn,
			configuration: cfg,
			logger:        logger.With("component", "filter"),
		},
		resender: reSender{
			in:     reSenderIn,
			out:    senderIn,
			events: make(map[string]Event),
		},
		sender: sender{
			in:            senderIn,
			configuration: cfg,
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
	go a.resender.Run(ctx, reSendInterval)
	go a.filter.Run(ctx)
	go a.ingressInformer.Run()
	defer a.ingressInformer.Cancel()
	<-ctx.Done()
}
