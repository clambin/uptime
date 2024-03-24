package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"strconv"
	"time"
)

var _ prometheus.Collector = HTTPMetrics{}

type HTTPMetrics struct {
	up         *prometheus.GaugeVec
	certExpiry *prometheus.GaugeVec
}

func NewHTTPMetrics(namespace, subsystem string, labels map[string]string) *HTTPMetrics {
	return &HTTPMetrics{
		up: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "up",
			Help:        "site is up/down",
			ConstLabels: labels,
		}, []string{"host"}),
		certExpiry: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "certificate_expiry_days",
			Help:        "number of days before the certificate expires",
			ConstLabels: labels,
		}, []string{"host"}),
	}
}

var bool2int = map[bool]int{
	true:  1,
	false: 0,
}

func (m HTTPMetrics) Observe(measurement HTTPMeasurement) {
	m.up.WithLabelValues(measurement.Host).Set(float64(bool2int[measurement.Up]))
	if measurement.IsTLS {
		m.certExpiry.WithLabelValues(measurement.Host).Set(measurement.TLSExpiry.Hours() / 24)
	}
}

func (m HTTPMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.up.Describe(ch)
	m.certExpiry.Describe(ch)
}

func (m HTTPMetrics) Collect(ch chan<- prometheus.Metric) {
	m.up.Collect(ch)
	m.certExpiry.Collect(ch)
}

var _ slog.LogValuer = HTTPMeasurement{}

type HTTPMeasurement struct {
	Host      string
	Up        bool
	Code      int
	Latency   time.Duration
	IsTLS     bool
	TLSExpiry time.Duration
}

func (m HTTPMeasurement) LogValue() slog.Value {
	attrs := make([]slog.Attr, 2, 5)
	attrs[0] = slog.String("target", m.Host)
	attrs[1] = slog.Bool("up", m.Up)
	if m.Code > 0 {
		attrs = append(attrs, slog.String("code", strconv.Itoa(m.Code)))
		attrs = append(attrs, slog.Duration("latency", m.Latency))
	}
	if m.IsTLS {
		attrs = append(attrs, slog.Duration("certExpiry", m.TLSExpiry))
	}
	return slog.GroupValue(attrs...)
}
