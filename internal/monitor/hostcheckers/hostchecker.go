package hostcheckers

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/clambin/uptime/internal/monitor/metrics"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type hostChecker struct {
	req        handlers.Request
	httpClient *http.Client
	metrics    HTTPObserver
	shutdown   chan struct{}
	logger     *slog.Logger
}

type HTTPObserver interface {
	Observe(httpMetrics metrics.HTTPMeasurement)
}

func newHostChecker(req handlers.Request, m HTTPObserver, c *http.Client, l *slog.Logger) *hostChecker {
	if c == nil {
		c = http.DefaultClient
	}
	return &hostChecker{
		req:        req,
		httpClient: c,
		metrics:    m,
		shutdown:   make(chan struct{}),
		logger:     l,
	}
}

func (h *hostChecker) Cancel() {
	h.shutdown <- struct{}{}
}

func (h *hostChecker) GetRequest() handlers.Request {
	return h.req
}

func (h *hostChecker) Run(interval time.Duration) {
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

func (h *hostChecker) ping() metrics.HTTPMeasurement {
	m := metrics.HTTPMeasurement{Host: h.req.Target}

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
