package agent

import (
	"context"
	"github.com/clambin/uptime/internal/monitor"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestSender_makeRequest(t *testing.T) {
	tests := []struct {
		name   string
		config Configuration
		event  Event
		want   monitor.Request
	}{
		{
			name:   "global",
			config: DefaultConfiguration,
			event:  Event{Type: AddEvent, Host: "foo"},
			want: monitor.Request{
				Target:    "foo",
				Method:    DefaultGlobalConfiguration.Method,
				ValidCode: DefaultGlobalConfiguration.ValidStatusCodes,
				Interval:  DefaultGlobalConfiguration.Interval,
			},
		},
		{
			name: "method",
			config: Configuration{
				Global: DefaultGlobalConfiguration,
				Hosts: map[string]EndpointConfiguration{
					"foo": {Method: http.MethodHead},
				},
			},
			event: Event{Type: AddEvent, Host: "foo"},
			want: monitor.Request{
				Target:    "foo",
				Method:    http.MethodHead,
				ValidCode: DefaultGlobalConfiguration.ValidStatusCodes,
				Interval:  DefaultGlobalConfiguration.Interval,
			},
		},
		{
			name: "interval",
			config: Configuration{
				Global: DefaultGlobalConfiguration,
				Hosts: map[string]EndpointConfiguration{
					"foo": {Interval: time.Minute},
				},
			},
			event: Event{Type: AddEvent, Host: "foo"},
			want: monitor.Request{
				Target:    "foo",
				Method:    DefaultGlobalConfiguration.Method,
				ValidCode: DefaultGlobalConfiguration.ValidStatusCodes,
				Interval:  time.Minute,
			},
		},
		{
			name: "status codes",
			config: Configuration{
				Global: DefaultGlobalConfiguration,
				Hosts: map[string]EndpointConfiguration{
					"foo": {ValidStatusCodes: []int{http.StatusUnauthorized}},
				},
			},
			event: Event{Type: AddEvent, Host: "foo"},
			want: monitor.Request{
				Target:    "foo",
				Method:    DefaultGlobalConfiguration.Method,
				ValidCode: []int{http.StatusUnauthorized},
				Interval:  DefaultGlobalConfiguration.Interval,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := sender{configuration: tt.config}
			assert.Equal(t, tt.want, s.makeRequest(tt.event))
		})
	}
}

func TestSender_Run(t *testing.T) {
	h := server{hosts: make(map[string]bool)}
	s := httptest.NewServer(&h)

	ch := make(chan Event)
	c := sender{
		in:            ch,
		configuration: DefaultConfiguration,
		httpClient:    http.DefaultClient,
		logger:        slog.Default(),
	}
	c.configuration.Monitor = s.URL

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)

	_, ok := h.getHost("foo")
	assert.False(t, ok)

	ch <- Event{Type: AddEvent, Host: "foo"}
	assert.Eventually(t, func() bool {
		up, ok := h.getHost("foo")
		return up && ok
	}, time.Second, time.Millisecond)

	ch <- Event{Type: DeleteEvent, Host: "foo"}
	assert.Eventually(t, func() bool {
		up, ok := h.getHost("foo")
		return !up && ok
	}, time.Second, time.Millisecond)

	s.Close()

	ch <- Event{Type: AddEvent, Host: "foo"}
	assert.Never(t, func() bool {
		up, ok := h.getHost("foo")
		return up && ok
	}, time.Second, time.Millisecond)
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
