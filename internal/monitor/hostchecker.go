package monitor

import (
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type HostChecker struct {
	req        Request
	httpClient *http.Client
	metrics    HTTPObserver
	shutdown   chan struct{}
	logger     *slog.Logger
}

type HTTPObserver interface {
	Observe(httpMetrics HTTPMeasurement)
}

func NewHostChecker(req Request, m HTTPObserver, c *http.Client, l *slog.Logger) *HostChecker {
	if c == nil {
		c = http.DefaultClient
	}
	return &HostChecker{
		req:        req,
		httpClient: c,
		metrics:    m,
		shutdown:   make(chan struct{}),
		logger:     l,
	}
}

func (h *HostChecker) Cancel() {
	h.shutdown <- struct{}{}
}

func (h *HostChecker) Run(interval time.Duration) {
	h.logger.Debug("hostchecker started", "request", h.req)
	defer h.logger.Debug("hostchecker stopped", "target", h.req.Target)
	for {
		h.metrics.Observe(h.ping())
		select {
		case <-h.shutdown:
			return
		case <-time.After(interval):
		}
	}
}

func (h *HostChecker) ping() HTTPMeasurement {
	m := HTTPMeasurement{Host: h.req.Target}

	target := h.req.Target
	if !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "http://") {
		target = "https://" + target
	}

	req, _ := http.NewRequest(h.req.Method, target, nil)

	start := time.Now()
	resp, err := h.httpClient.Do(req)

	if err != nil {
		h.logger.Debug("measurement failed", "err", err)
		return m
	}

	_ = resp.Body.Close()
	m.Up = h.req.ValidCodes.Contains(resp.StatusCode)
	m.Code = resp.StatusCode
	m.Latency = time.Since(start)
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		m.IsTLS = true
		// PeerCertificates: the first one in the list is the leaf certificate
		m.TLSExpiry = time.Until(resp.TLS.PeerCertificates[0].NotAfter)
	}

	h.logger.Debug("measurement made", "up", m.Up, "latency", m.Latency, "code", m.Code)
	return m
}
