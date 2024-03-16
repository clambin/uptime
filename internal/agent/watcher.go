package agent

import (
	"github.com/clambin/go-common/set"
	netv1 "k8s.io/api/networking/v1"
	"log/slog"
	"strings"
)

type ingressWatcher struct {
	out     chan<- event
	metrics *Metrics
	logger  *slog.Logger
}

func (w ingressWatcher) OnAdd(ingress any, _ bool) {
	w.send(event{eventType: addEvent, ingress: ingress.(*netv1.Ingress)})
}

func (w ingressWatcher) OnUpdate(oldIngress, newIngress any) {
	oldEv := event{eventType: deleteEvent, ingress: oldIngress.(*netv1.Ingress)}
	newEv := event{eventType: addEvent, ingress: newIngress.(*netv1.Ingress)}

	oldHostnames := set.New(oldEv.targetHosts()...)
	newHostnames := set.New(newEv.targetHosts()...)

	if strings.Join(oldHostnames.ListOrdered(), ",") != strings.Join(newHostnames.ListOrdered(), ",") {
		w.send(oldEv)
		w.send(newEv)
	}
}

func (w ingressWatcher) OnDelete(ingress any) {
	w.send(event{eventType: deleteEvent, ingress: ingress.(*netv1.Ingress)})
}

func (w ingressWatcher) send(ev event) {
	w.logger.Debug("ingress change detected", "event", ev)
	w.out <- ev
	if w.metrics != nil {
		w.metrics.ObserveEvent(ev)
	}
}
