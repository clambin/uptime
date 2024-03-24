package monitor

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/clambin/uptime/internal/monitor/metrics"
	"net/http"
	"time"
)

const DefaultClientTimeout = 10 * time.Second

func New(metrics *metrics.HostMetrics, httpClient *http.Client) http.Handler {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: DefaultClientTimeout,
		}
	}

	hc := hostCheckers{
		metrics:      metrics,
		httpClient:   httpClient,
		hostCheckers: make(map[string]*hostChecker),
	}
	h := http.NewServeMux()
	h.Handle("/target", handlers.TargetHandler{TargetManager: &hc})
	return h
}
