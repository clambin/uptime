package agent

import (
	"context"
	"fmt"
	"github.com/clambin/uptime/internal/agent/informer"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http"
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
	filter          Filter
	sender          Sender
}

func New(c *kubernetes.Clientset, monitor string, token string, logger *slog.Logger) (*Agent, error) {
	filterIn := make(chan Event)
	filterOut := make(chan Event)
	i, err := informer.NewIngressInformer(c, v1.NamespaceAll, &IngressWatcher{
		WatcherOut: filterIn,
		logger:     logger.With("component", "watcher"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes informer: %w", err)
	}
	return &Agent{
		ingressInformer: i,
		filter: Filter{
			EventsIn:  filterIn,
			EventsOut: filterOut,
		},
		sender: Sender{
			Process:    filterOut,
			monitor:    monitor,
			token:      token,
			httpClient: http.DefaultClient,
			logger:     logger.With("component", "sender"),
		},
	}, nil
}

func (a *Agent) Run(ctx context.Context) {
	go a.sender.Run(ctx)
	go a.filter.Run(ctx)
	go a.ingressInformer.Run()
	defer a.ingressInformer.Cancel()
	<-ctx.Done()
}
