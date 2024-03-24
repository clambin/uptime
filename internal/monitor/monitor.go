package monitor

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	"net/http"
	"time"
)

const DefaultClientTimeout = 10 * time.Second

func New(metrics *HostMetrics, httpClient *http.Client) http.Handler {
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
