package monitor

import (
	"github.com/clambin/uptime/pkg/logger"
	"net/http"
)

var _ http.Handler = &Monitor{}

type Monitor struct {
	http.Handler
	metrics      *HTTPMetrics
	httpClient   *http.Client
	hostCheckers *hostCheckers
}

func New(metrics *HTTPMetrics, httpClient *http.Client) *Monitor {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	m := Monitor{
		metrics:      metrics,
		httpClient:   httpClient,
		hostCheckers: &hostCheckers{hostCheckers: make(map[string]checker)},
	}
	m.Handler = m.makeHandler()

	return &m
}

func (m *Monitor) makeHandler() http.Handler {
	h := http.NewServeMux()
	h.HandleFunc("POST /target", m.addTarget)
	h.HandleFunc("DELETE /target", m.removeTarget)

	return h
}

func (m *Monitor) addTarget(w http.ResponseWriter, r *http.Request) {
	l := logger.Logger(r)
	req, err := ParseRequest(r)
	if err != nil {
		l.Error("invalid request", "err", err)
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	h := newHostChecker(req, m.metrics, m.httpClient, l.With("target", req.Target))
	m.hostCheckers.add(req.Target, h, req.Interval)

	l.Info("target added", "req", req)
	w.WriteHeader(http.StatusOK)
}

func (m *Monitor) removeTarget(w http.ResponseWriter, r *http.Request) {
	l := logger.Logger(r)

	req, err := ParseRequest(r)
	if err != nil {
		l.Error("invalid request", "err", err)
		w.Header().Set("Content-Type", "plain/text")
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	m.hostCheckers.remove(req.Target)
	l.Debug("target removed", "target", req.Target)
	w.WriteHeader(http.StatusOK)
}
