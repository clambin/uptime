package agent

import (
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"testing"
)

func TestIngressWatcher(t *testing.T) {
	ch := make(chan Event)
	w := ingressWatcher{
		out:       ch,
		logger:    slog.Default(),
		hostnames: make(map[string]string),
	}

	ingress := netv1.Ingress{
		ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "bar", Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint}},
		Spec:       netv1.IngressSpec{Rules: []netv1.IngressRule{{Host: "example.com"}}},
	}

	go w.OnAdd(&ingress, true)
	ev := <-ch
	assert.Equal(t, AddEvent, ev.Type)

	go w.OnDelete(&ingress)
	ev = <-ch
	assert.Equal(t, DeleteEvent, ev.Type)

	ingress2 := ingress
	ingress2.Spec.Rules = []netv1.IngressRule{{Host: "example.com/foo"}}

	go w.OnUpdate(&ingress, &ingress2)
	assert.Equal(t, DeleteEvent, (<-ch).Type)
	assert.Equal(t, AddEvent, (<-ch).Type)
}
