package monitor

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/clambin/uptime/internal/monitor/hostcheckers"
	"github.com/clambin/uptime/internal/monitor/metrics"
	"net/http"
	"time"
)

const DefaultClientTimeout = 10 * time.Second

func New(metrics *metrics.HostMetrics, httpClient *http.Client) http.Handler {
	h := http.NewServeMux()
	h.Handle("/target", handlers.TargetHandler{TargetManager: hostcheckers.New(metrics, httpClient)})
	return h
}
