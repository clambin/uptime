package agent

import (
	netv1 "k8s.io/api/networking/v1"
	"log/slog"
)

type IngressWatcher struct {
	WatcherOut chan<- Event
	logger     *slog.Logger
}

func (w IngressWatcher) OnAdd(ingress any, _ bool) {
	w.send(AddEvent, ingress.(*netv1.Ingress))
}

func (w IngressWatcher) OnUpdate(oldIngress, newIngress any) {
	w.send(DeleteEvent, oldIngress.(*netv1.Ingress))
	w.send(AddEvent, newIngress.(*netv1.Ingress))
}

func (w IngressWatcher) OnDelete(ingress any) {
	w.send(DeleteEvent, ingress.(*netv1.Ingress))
}

func (w IngressWatcher) send(eventType EventType, ingress *netv1.Ingress) {
	for i := range ingress.Spec.Rules {
		w.logger.Debug("ingress found", "name", ingress.Name, "namespace", ingress.Namespace, "host", ingress.Spec.Rules[i].Host, "event", eventType)
		w.WatcherOut <- Event{
			Type:        eventType,
			Host:        ingress.Spec.Rules[i].Host,
			Annotations: ingress.ObjectMeta.Annotations,
		}
	}
}
