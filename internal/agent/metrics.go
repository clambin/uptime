package agent

import (
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

var _ prometheus.Collector = Metrics{}

type Metrics struct {
	IngressEvents *prometheus.CounterVec
}

func NewMetrics(namespace, subsystem string, labels map[string]string) *Metrics {
	return &Metrics{
		IngressEvents: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "ingress_events_count",
			Help:        "number of ingress events received from kubernetes",
			ConstLabels: labels,
		}, []string{"name", "namespace", "type"}),
	}
}

func (m Metrics) ObserveEvent(ev event) {
	m.IngressEvents.WithLabelValues(ev.name(), ev.namespace(), strings.ToLower(string(ev.eventType))).Add(1)
}

func (m Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.IngressEvents.Describe(ch)
}

func (m Metrics) Collect(ch chan<- prometheus.Metric) {
	m.IngressEvents.Collect(ch)
}
