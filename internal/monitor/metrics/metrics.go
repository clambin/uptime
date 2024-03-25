package metrics

import (
	"github.com/clambin/go-common/http/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var _ prometheus.Collector = HostMetrics{}

type HostMetrics struct {
	up         *prometheus.GaugeVec
	certExpiry *prometheus.GaugeVec
}

func NewHostMetrics(namespace, subsystem string, labels map[string]string) *HostMetrics {
	return &HostMetrics{
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

func (m HostMetrics) Observe(measurement HTTPMeasurement) {
	m.up.WithLabelValues(measurement.Host).Set(float64(bool2int[measurement.Up]))
	if measurement.IsTLS {
		m.certExpiry.WithLabelValues(measurement.Host).Set(measurement.TLSExpiry.Hours() / 24)
	}
}

func (m HostMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.up.Describe(ch)
	m.certExpiry.Describe(ch)
}

func (m HostMetrics) Collect(ch chan<- prometheus.Metric) {
	m.up.Collect(ch)
	m.certExpiry.Collect(ch)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ metrics.RequestMetrics = &httpMetrics{}

type httpMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func NewHTTPMetrics(namespace, subsystem string, labels map[string]string, buckets ...float64) metrics.RequestMetrics {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return &httpMetrics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        metrics.RequestTotal,
			Help:        "total number of http requests",
			ConstLabels: labels,
		},
			[]string{"host", "code"},
		),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        metrics.RequestsDuration,
			Help:        "duration of http requests",
			Buckets:     buckets,
			ConstLabels: labels,
		},
			[]string{"host"},
		),
	}
}

func (h httpMetrics) Measure(req *http.Request, statusCode int, duration time.Duration) {
	code := strconv.Itoa(statusCode)
	h.requests.WithLabelValues(req.URL.Hostname(), code).Inc()
	h.duration.WithLabelValues(req.URL.Hostname()).Observe(duration.Seconds())
}

func (h httpMetrics) Describe(ch chan<- *prometheus.Desc) {
	h.requests.Describe(ch)
	h.duration.Describe(ch)
}

func (h httpMetrics) Collect(ch chan<- prometheus.Metric) {
	h.requests.Collect(ch)
	h.duration.Collect(ch)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

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
