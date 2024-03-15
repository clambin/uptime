package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fcache "k8s.io/client-go/tools/cache/testing"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAgent_Run(t *testing.T) {
	h := server{hosts: make(map[string]bool)}
	s := httptest.NewServer(&h)
	defer s.Close()

	l := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := DefaultConfiguration
	cfg.Monitor = s.URL

	f := fcache.NewFakeControllerSource()

	a, err := NewWithListWatcher(f, cfg, NewMetrics("uptime", "agent"), l)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.Run(ctx)

	f.Add(&netv1.Ingress{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "default", Annotations: map[string]string{
			traefikEndpointAnnotation: traefikExternalEndpoint,
		}},
		Spec: netv1.IngressSpec{Rules: []netv1.IngressRule{
			{Host: "https://foo.example.com"},
		}},
		Status: netv1.IngressStatus{},
	})

	assert.Eventually(t, func() bool {
		up, ok := h.getHost("https://foo.example.com")
		return ok && up
	}, 5*time.Second, time.Second)
}

func BenchmarkAgent(b *testing.B) {
	filterIn := make(chan Event)
	resenderIn := make(chan Event)
	resenderOut := make(chan Event, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := ingressWatcher{
		out:       filterIn,
		logger:    slog.Default(),
		hostnames: make(map[string]string),
	}
	f := filter{
		in:  filterIn,
		out: resenderIn,
	}
	for range 5 {
		go f.Run(ctx)
	}
	r := reSender{
		in:     resenderIn,
		out:    resenderOut,
		events: make(map[string]Event),
	}
	go r.Run(ctx, time.Hour)

	i := netv1.Ingress{
		ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "bar", Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint}},
		Spec:       netv1.IngressSpec{Rules: []netv1.IngressRule{{Host: "foo"}}},
		Status:     netv1.IngressStatus{},
	}

	b.ResetTimer()
	go func() {
		for range b.N {
			w.OnAdd(&i, false)
		}
	}()

	for range b.N {
		<-resenderOut
	}
}
