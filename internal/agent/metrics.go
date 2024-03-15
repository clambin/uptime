package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
	"time"
)

var _ prometheus.Collector = Metrics{}

type Metrics struct {
	IngressEvents *prometheus.CounterVec
	Requests      *prometheus.CounterVec
	Latency       *prometheus.HistogramVec
}

func NewMetrics(namespace, subsystem string) *Metrics {
	return &Metrics{
		IngressEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "ingress_events",
			Help:        "number of ingress events received from kubernetes",
			ConstLabels: nil,
		}, []string{"host", "type"}),
		Requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "requests",
			Help:        "number of requests sent to the monitor",
			ConstLabels: nil,
		}, []string{"code"}),
		Latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_latency",
			Help:      "request latency latency",
			Buckets:   []float64{0.1, 0.2, 0.5, 1, 2, 5},
		}, []string{"code"}),
	}
}

func (m Metrics) ObserveEvent(ev Event) {
	m.IngressEvents.WithLabelValues(ev.Host, strings.ToLower(string(ev.Type))).Add(1)
}

func (m Metrics) ObserveRequest(code int, latency time.Duration) {
	codeString := strconv.Itoa(code)
	m.Requests.WithLabelValues(codeString).Add(1)
	m.Latency.WithLabelValues(codeString).Observe(latency.Seconds())
}

func (m Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.IngressEvents.Describe(ch)
	m.Requests.Describe(ch)
	m.Latency.Describe(ch)
}

func (m Metrics) Collect(ch chan<- prometheus.Metric) {
	m.IngressEvents.Collect(ch)
	m.Requests.Collect(ch)
	m.Latency.Collect(ch)
}
