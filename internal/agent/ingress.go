package agent

import (
	"github.com/clambin/go-common/set"
	netv1 "k8s.io/api/networking/v1"
	"log/slog"
	"strings"
)

type IngressWatcher struct {
	WatcherOut chan<- Event
	logger     *slog.Logger
	hostnames  map[string]string
}

func (w IngressWatcher) OnAdd(ingress any, _ bool) {
	w.send(AddEvent, ingress.(*netv1.Ingress))
}

func (w IngressWatcher) OnUpdate(oldIngress, newIngress any) {
	oldHostnames := set.New(getHostnames(oldIngress.(*netv1.Ingress))...)
	newHostnames := set.New(getHostnames(newIngress.(*netv1.Ingress))...)

	if strings.Join(oldHostnames.ListOrdered(), ",") == strings.Join(newHostnames.ListOrdered(), ",") {
		return
	}

	w.send(DeleteEvent, oldIngress.(*netv1.Ingress))
	w.send(AddEvent, newIngress.(*netv1.Ingress))
}

func (w IngressWatcher) OnDelete(ingress any) {
	w.send(DeleteEvent, ingress.(*netv1.Ingress))
}

func (w IngressWatcher) send(eventType EventType, ingress *netv1.Ingress) {
	for _, hostname := range getHostnames(ingress) {
		w.logger.Debug("ingress detected", "name", ingress.Name, "namespace", ingress.Namespace, "host", hostname, "event", eventType)
		w.WatcherOut <- Event{
			Type:        eventType,
			Host:        hostname,
			Annotations: ingress.ObjectMeta.Annotations,
		}
	}
}

func getHostnames(ingress *netv1.Ingress) []string {
	hostnames := make([]string, len(ingress.Spec.Rules))
	for i := range ingress.Spec.Rules {
		hostname := ingress.Spec.Rules[i].Host
		if !strings.HasPrefix("https://", hostname) {
			hostname = "https://" + hostname
		}
		hostnames[i] = hostname
	}
	return hostnames
}
