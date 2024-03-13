package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"time"
)

var _ prometheus.Collector = HTTPMetrics{}

type HTTPMetrics struct {
	Latency    *prometheus.HistogramVec
	Up         *prometheus.GaugeVec
	CertExpiry *prometheus.GaugeVec
}

func NewHTTPMetrics(namespace, subsystem string) *HTTPMetrics {
	return &HTTPMetrics{
		Latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "latency",
			Help:      "site latency",
			Buckets:   []float64{0.25, 0.5, 0.75, 1, 2, 4},
		}, []string{"host", "code"}),
		Up: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "up",
			Help:      "site is up/down",
		}, []string{"host"}),
		CertExpiry: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "_certificate_expiry_days",
			Help:      "number of days before the certificate expires",
		}, []string{"host"}),
	}
}

func (m HTTPMetrics) Observe(measurement HTTPMeasurement) {
	m.Latency.WithLabelValues(measurement.Host, measurement.Code).Observe(measurement.Latency.Seconds())
	var v float64
	if measurement.Up {
		v = 1
	}
	m.Up.WithLabelValues(measurement.Host).Set(v)
	if measurement.IsTLS {
		m.CertExpiry.WithLabelValues(measurement.Host).Set(measurement.TLSExpiry.Hours() / 24)
	}
}
func (m HTTPMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.Up.Describe(ch)
	m.Latency.Describe(ch)
	m.CertExpiry.Describe(ch)
}

func (m HTTPMetrics) Collect(ch chan<- prometheus.Metric) {
	m.Up.Collect(ch)
	m.Latency.Collect(ch)
	m.CertExpiry.Collect(ch)
}

var _ slog.LogValuer = HTTPMeasurement{}

type HTTPMeasurement struct {
	Host      string
	Up        bool
	Code      string
	Latency   time.Duration
	IsTLS     bool
	TLSExpiry time.Duration
}

func (m HTTPMeasurement) LogValue() slog.Value {
	attrs := make([]slog.Attr, 2, 5)
	attrs[0] = slog.String("target", m.Host)
	attrs[1] = slog.Bool("up", m.Up)
	if m.Code != "" {
		attrs = append(attrs, slog.String("code", m.Code))
		attrs = append(attrs, slog.Duration("latency", m.Latency))
	}
	if m.IsTLS {
		attrs = append(attrs, slog.Duration("certExpiry", m.TLSExpiry))
	}
	return slog.GroupValue(attrs...)
}
