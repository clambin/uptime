package agent

import (
	netv1 "k8s.io/api/networking/v1"
	"log/slog"
)

type eventType string

const (
	addEvent    eventType = "ADD"
	deleteEvent eventType = "DELETE"
)

var _ slog.LogValuer = event{}

type event struct {
	eventType eventType
	ingress   *netv1.Ingress
}

func (e event) name() string {
	return e.ingress.Name
}

func (e event) namespace() string {
	return e.ingress.Namespace
}

func (e event) hasAnnotation(annotation, value string) bool {
	v, ok := e.ingress.Annotations[annotation]
	return ok && v == value
}

func (e event) targetHosts() []string {
	targets := make([]string, len(e.ingress.Spec.Rules))
	for i := range e.ingress.Spec.Rules {
		targets[i] = e.ingress.Spec.Rules[i].Host
	}
	return targets
}

func (e event) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("type", string(e.eventType)),
		slog.String("name", e.name()),
		slog.String("namespace", e.namespace()),
	)
}
