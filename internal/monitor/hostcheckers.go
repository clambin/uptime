package monitor

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	"log/slog"
	"net/http"
	"sync"
)

type hostCheckers struct {
	metrics      *HostMetrics
	httpClient   *http.Client
	lock         sync.Mutex
	hostCheckers map[string]*hostChecker
}

func (h *hostCheckers) Add(request handlers.Request, logger *slog.Logger) {
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
	hc := newHostChecker(request, h.metrics, h.httpClient, logger.With("target", request.Target))
	h.hostCheckers[request.Target] = hc
	go hc.Run(request.Interval)
}

func (h *hostCheckers) Remove(request handlers.Request, logger *slog.Logger) {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[request.Target]
	if ok {
		logger.Info("target removed", "target", request)
		c.Cancel()
		delete(h.hostCheckers, request.Target)
	}
}
