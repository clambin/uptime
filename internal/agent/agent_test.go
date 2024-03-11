package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"log/slog"
	"net/http/httptest"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"testing"
	"time"
)

func TestAgent_Run(t *testing.T) {
	c, err := kubernetes.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		t.Skip("not connected to cluster")
	}

	h := server{hosts: make(map[string]bool)}
	s := httptest.NewServer(&h)
	defer s.Close()

	l := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	a, err := New(c, s.URL, "", l)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.Run(ctx)

	assert.Eventually(t, func() bool {
		up, ok := h.getHost("plex.agrajag.duckdns.org")
		return ok && up
	}, 5*time.Second, time.Second)
}
