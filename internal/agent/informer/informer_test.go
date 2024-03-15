package informer_test

import (
	"github.com/clambin/uptime/internal/agent/informer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	fcache "k8s.io/client-go/tools/cache/testing"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sync/atomic"
	"testing"
	"time"
)

func TestInformer(t *testing.T) {
	watchlist := fcache.NewFakeControllerSource()
	w := watcher{t: t}
	i, err := informer.New(watchlist, time.Minute, new(netv1.Ingress), &w)
	require.NoError(t, err)

	go i.Run()
	defer i.Cancel()

	ingresses := []*netv1.Ingress{
		{
			TypeMeta:   v1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"},
			ObjectMeta: v1.ObjectMeta{Name: "foo", Namespace: "bar", Annotations: map[string]string{"traefik.ingress.kubernetes.io/router.entrypoints": "websecure"}},
			Spec:       netv1.IngressSpec{Rules: []netv1.IngressRule{{Host: "argocd.192.168.0.25.nip.io"}}},
		},
		{
			TypeMeta:   v1.TypeMeta{APIVersion: "networking.k8s.io/v1", Kind: "Ingress"},
			ObjectMeta: v1.ObjectMeta{Name: "snafu", Namespace: "bar", Annotations: map[string]string{"traefik.ingress.kubernetes.io/router.entrypoints": "websecure"}},
			Spec:       netv1.IngressSpec{Rules: []netv1.IngressRule{{Host: "argocd.192.168.0.25.nip.io"}}},
		},
	}

	for _, ingress := range ingresses {
		watchlist.Add(ingress)
	}

	assert.Eventually(t, func() bool {
		return i.ResourceEventHandlerRegistration.HasSynced()
	}, time.Second, time.Millisecond)
	assert.Equal(t, len(ingresses), int(w.added.Load()))

	for _, ingress := range ingresses {
		watchlist.Delete(ingress)
	}
}

func TestIngressInformer(t *testing.T) {
	t.Skip()
	c, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	require.NoError(t, err)

	w := watcher{t: t}
	i, err := informer.NewIngressInformer(c, v1.NamespaceAll, &w)
	require.NoError(t, err)

	go i.Run()
	defer i.Cancel()

	assert.Eventually(t, func() bool {
		return i.ResourceEventHandlerRegistration.HasSynced()
	}, time.Second, time.Millisecond)

	assert.NotZero(t, w.added.Load())
}

var _ cache.ResourceEventHandler = &watcher{}

type watcher struct {
	t     *testing.T
	added atomic.Int32
}

func (w *watcher) OnAdd(obj any, _ bool) {
	w.t.Helper()
	_, ok := obj.(*netv1.Ingress)
	assert.True(w.t, ok)
	w.added.Add(1)
}

func (w *watcher) OnUpdate(oldObj, newObj any) {
	w.t.Helper()
	w.OnDelete(oldObj)
	w.OnAdd(newObj, false)
}

func (w *watcher) OnDelete(obj any) {
	w.t.Helper()
	_, ok := obj.(*netv1.Ingress)
	assert.True(w.t, ok)
}
