package agent

import (
	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"testing"
)

func TestIngressWatcher(t *testing.T) {
	ch := make(chan event)
	w := ingressWatcher{
		out:    ch,
		logger: slog.Default(),
	}

	go w.OnAdd(&validIngress, true)
	ev := <-ch
	assert.Equal(t, addEvent, ev.eventType)

	go w.OnDelete(&validIngress)
	ev = <-ch
	assert.Equal(t, deleteEvent, ev.eventType)

	ingress2 := netv1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        "valid",
			Namespace:   "foo",
			Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint},
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{{Host: "example.com/foo"}},
		},
	}

	go w.OnUpdate(&validIngress, &ingress2)
	assert.Equal(t, deleteEvent, (<-ch).eventType)
	assert.Equal(t, addEvent, (<-ch).eventType)
}
