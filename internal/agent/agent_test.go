package agent

import (
	"context"
	"github.com/clambin/uptime/internal/monitor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	netv1 "k8s.io/api/networking/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fcache "k8s.io/client-go/tools/cache/testing"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

var (
	validIngress = netv1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:        "valid",
			Namespace:   "foo",
			Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint},
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{{Host: "example.com"}},
		},
	}
	invalidIngress = netv1.Ingress{
		ObjectMeta: v1.ObjectMeta{
			Name:      "invalid",
			Namespace: "bar",
		},
		Spec: netv1.IngressSpec{
			Rules: []netv1.IngressRule{{Host: "example.com"}},
		},
	}
)

func TestAgent_Run(t *testing.T) {
	h := server{hosts: make(map[string]bool)}
	s := httptest.NewServer(&h)
	defer s.Close()

	l := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := DefaultConfiguration
	cfg.Monitor = s.URL

	f := fcache.NewFakeControllerSource()
	m := NewMetrics()

	_, err := NewWithListWatcher(f, Configuration{}, m, l)
	assert.Error(t, err)

	a, err := NewWithListWatcher(f, cfg, NewMetrics(), l)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.Run(ctx)

	f.Add(&validIngress)

	assert.Eventually(t, func() bool {
		up, ok := h.getHost("example.com")
		return ok && up
	}, 5*time.Second, time.Second)
}

func BenchmarkAgent(b *testing.B) {
	filterIn := make(chan event)
	resenderIn := make(chan event)
	resenderOut := make(chan event, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w := ingressWatcher{
		out:    filterIn,
		logger: slog.Default(),
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
		events: make(map[string]event),
	}
	go r.Run(ctx, time.Hour)

	b.ResetTimer()
	go func() {
		for range b.N {
			w.OnAdd(&validIngress, false)
		}
	}()

	for range b.N {
		<-resenderOut
	}
}

type server struct {
	lock  sync.Mutex
	hosts map[string]bool
}

func (s *server) getHost(target string) (bool, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	up, ok := s.hosts[target]
	return up, ok
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := monitor.ParseRequest(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	switch r.Method {
	case http.MethodPost:
		s.hosts[req.Target] = true
	case http.MethodDelete:
		s.hosts[req.Target] = false
	default:
		http.Error(w, "", http.StatusMethodNotAllowed)
		return
	}
}
