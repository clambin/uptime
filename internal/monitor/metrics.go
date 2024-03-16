package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"strconv"
	"time"
)

var _ prometheus.Collector = HTTPMetrics{}

type HTTPMetrics struct {
	Latency    *prometheus.HistogramVec
	Up         *prometheus.GaugeVec
	CertExpiry *prometheus.GaugeVec
}

func NewHTTPMetrics() *HTTPMetrics {
	return &HTTPMetrics{
		Latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "uptime",
			Subsystem: "monitor",
			Name:      "latency",
			Help:      "site latency",
			Buckets:   []float64{0.25, 0.5, 0.75, 1, 2, 4},
		}, []string{"host", "code"}),
		Up: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "uptime",
			Subsystem: "monitor",
			Name:      "up",
			Help:      "site is up/down",
		}, []string{"host"}),
		CertExpiry: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "uptime",
			Subsystem: "monitor",
			Name:      "certificate_expiry_days",
			Help:      "number of days before the certificate expires",
		}, []string{"host"}),
	}
}

var bool2int = map[bool]int{
	true:  1,
	false: 0,
}

func (m HTTPMetrics) Observe(measurement HTTPMeasurement) {
	m.Up.WithLabelValues(measurement.Host).Set(float64(bool2int[measurement.Up]))
	if measurement.Code > 0 {
		m.Latency.WithLabelValues(measurement.Host, strconv.Itoa(measurement.Code)).Observe(measurement.Latency.Seconds())
	}
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
