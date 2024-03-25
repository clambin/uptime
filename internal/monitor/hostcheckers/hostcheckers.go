package hostcheckers

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	metrics2 "github.com/clambin/uptime/internal/monitor/metrics"
	"log/slog"
	"net/http"
	"sync"
)

type HostCheckers struct {
	Metrics      *metrics2.HostMetrics
	HTTPClient   *http.Client
	lock         sync.Mutex
	hostCheckers map[string]*hostChecker
}

func New(hostMetrics *metrics2.HostMetrics, httpClient *http.Client) *HostCheckers {
	return &HostCheckers{
		Metrics:      hostMetrics,
		HTTPClient:   httpClient,
		hostCheckers: make(map[string]*hostChecker),
	}
}

func (h *HostCheckers) Add(request handlers.Request, logger *slog.Logger) {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[request.Target]
	if ok {
		if c.GetRequest().Equals(request) {
			return
		}
		logger.Debug("target replaced. shutting down old hostChecker", "target", request.Target)
		c.Cancel()
		delete(h.hostCheckers, request.Target)
	}

	logger.Info("target added", "target", request)
	hc := newHostChecker(request, h.Metrics, h.HTTPClient, logger.With("target", request.Target))
	h.hostCheckers[request.Target] = hc
	go hc.Run(request.Interval)
}

func (h *HostCheckers) Remove(request handlers.Request, logger *slog.Logger) {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[request.Target]
	if ok {
		logger.Info("target removed", "target", request)
		c.Cancel()
		delete(h.hostCheckers, request.Target)
	}
}
