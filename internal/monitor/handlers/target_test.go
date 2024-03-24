package handlers_test

import (
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestTargetHandler(t *testing.T) {
	m := mgr{target: make(map[string]struct{})}
	h := handlers.TargetHandler{
		TargetManager: &m,
	}

	req, _ := http.NewRequest(http.MethodHead, "/?target=localhost:8080", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	req, _ = http.NewRequest(http.MethodPost, "/?target=localhost:8080", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, m.has("localhost:8080"))

	req, _ = http.NewRequest(http.MethodDelete, "/?target=localhost:8080", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.False(t, m.has("localhost:8080"))

	req, _ = http.NewRequest(http.MethodDelete, "/", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

var _ handlers.TargetManager = &mgr{}

type mgr struct {
	target map[string]struct{}
	lock   sync.Mutex
}

func (m *mgr) Add(request handlers.Request, _ *slog.Logger) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.target[request.Target] = struct{}{}
}

func (m *mgr) Remove(request handlers.Request) {
	m.lock.Lock()
	defer m.lock.Unlock()
	delete(m.target, request.Target)
}

func (m *mgr) has(target string) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	_, ok := m.target[target]
	return ok
}
