package monitor

import (
	"fmt"
	"github.com/clambin/go-common/set"
	"log/slog"
	"net/http"
	"time"
)

type HostChecker struct {
	target     string
	method     string
	httpClient *http.Client
	metrics    HTTPObserver
	validCodes set.Set[int]
	shutdown   chan struct{}
	logger     *slog.Logger
}

type HTTPObserver interface {
	Observe(httpMetrics HTTPMeasurement)
}

func NewHostChecker(target string, method string, m HTTPObserver, c *http.Client, l *slog.Logger, validCodes ...int) *HostChecker {
	if c == nil {
		c = http.DefaultClient
	}
	if len(validCodes) == 0 {
		validCodes = []int{http.StatusOK}
	}
	return &HostChecker{
		target:     target,
		method:     method,
		validCodes: set.New(validCodes...),
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
	h.logger.Debug("hostchecker started", "target", h.target, "method", h.method, "codes", h.validCodes)
	defer h.logger.Debug("hostchecker stopped", "target", h.target)
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
	m := HTTPMeasurement{Host: h.target}

	req, _ := http.NewRequest(h.method, h.target, nil)

	start := time.Now()
	resp, err := h.httpClient.Do(req)

	if err != nil {
		h.logger.Debug("measurement failed", "result", m, "err", err)
		return m
	}

	_ = resp.Body.Close()
	m.Code = fmt.Sprintf("%03d", resp.StatusCode)
	if m.Up = h.validCodes.Contains(resp.StatusCode); m.Up {
		m.Latency = time.Since(start)
	}
	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		m.IsTLS = true
		m.TLSExpiry = time.Until(resp.TLS.PeerCertificates[0].NotAfter)
	}

	h.logger.Debug("measurement made", "result", m)
	return m
}
