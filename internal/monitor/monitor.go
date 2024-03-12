package monitor

import (
	"github.com/clambin/uptime/pkg/logger"
	"net/http"
	"sync"
)

var _ http.Handler = &Monitor{}

type Monitor struct {
	http.Handler
	metrics      *HTTPMetrics
	httpClient   *http.Client
	lock         sync.Mutex
	hostCheckers HostCheckers
}

func New(metrics *HTTPMetrics, httpClient *http.Client) *Monitor {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	m := Monitor{
		metrics:      metrics,
		httpClient:   httpClient,
		hostCheckers: make(HostCheckers),
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

	l.Debug("adding target", "req", req)

	m.lock.Lock()
	defer m.lock.Unlock()

	h := NewHostChecker(req.Target, req.Method, m.metrics, m.httpClient, l.With("target", req.Target), req.ValidCode...)
	m.hostCheckers.Add(req.Target, h, req.Interval)
	//w.WriteHeader(http.StatusOK)
}

func (m *Monitor) removeTarget(w http.ResponseWriter, r *http.Request) {
	l := logger.Logger(r)

	req, err := ParseRequest(r)
	if err != nil {
		l.Error("invalid request", "err", err)
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	l.Debug("removing target", "target", req.Target)

	if !m.hostCheckers.Remove(req.Target) {
		http.Error(w, "no checker running for "+req.Target, http.StatusNotFound)
	}
	w.WriteHeader(http.StatusOK)
}
