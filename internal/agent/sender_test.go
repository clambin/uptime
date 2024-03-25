package agent

import (
	"context"
	"github.com/clambin/go-common/set"
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSender_makeRequests(t *testing.T) {
	tests := []struct {
		name   string
		config Configuration
		event  event
		want   []handlers.Request
	}{
		{
			name:   "global",
			config: DefaultConfiguration,
			event:  event{eventType: addEvent, ingress: &validIngress},
			want: []handlers.Request{{
				Target:     "example.com",
				Method:     DefaultGlobalConfiguration.Method,
				ValidCodes: set.New(DefaultGlobalConfiguration.ValidStatusCodes...),
				Interval:   DefaultGlobalConfiguration.Interval,
			}},
		},
		{
			name: "method",
			config: Configuration{
				Global: DefaultGlobalConfiguration,
				Hosts: map[string]EndpointConfiguration{
					"example.com": {Method: http.MethodHead},
				},
			},
			event: event{eventType: addEvent, ingress: &validIngress},
			want: []handlers.Request{{
				Target:     "example.com",
				Method:     http.MethodHead,
				ValidCodes: set.New(DefaultGlobalConfiguration.ValidStatusCodes...),
				Interval:   DefaultGlobalConfiguration.Interval,
			}},
		},
		{
			name: "interval",
			config: Configuration{
				Global: DefaultGlobalConfiguration,
				Hosts: map[string]EndpointConfiguration{
					"example.com": {Interval: time.Minute},
				},
			},
			event: event{eventType: addEvent, ingress: &validIngress},
			want: []handlers.Request{{
				Target:     "example.com",
				Method:     DefaultGlobalConfiguration.Method,
				ValidCodes: set.New(DefaultGlobalConfiguration.ValidStatusCodes...),
				Interval:   time.Minute,
			}},
		},
		{
			name: "status codes",
			config: Configuration{
				Global: DefaultGlobalConfiguration,
				Hosts: map[string]EndpointConfiguration{
					"example.com": {ValidStatusCodes: []int{http.StatusUnauthorized}},
				},
			},
			event: event{eventType: addEvent, ingress: &validIngress},
			want: []handlers.Request{{
				Target:     "example.com",
				Method:     DefaultGlobalConfiguration.Method,
				ValidCodes: set.New(http.StatusUnauthorized),
				Interval:   DefaultGlobalConfiguration.Interval,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := sender{configuration: tt.config}
			assert.Equal(t, tt.want, s.makeRequests(tt.event))
		})
	}
}

func TestSender_Run(t *testing.T) {
	h := server{hosts: make(map[string]bool)}
	s := httptest.NewServer(&h)

	ch := make(chan event)
	c := sender{
		in:            ch,
		configuration: DefaultConfiguration,
		httpClient:    http.DefaultClient,
		logger:        slog.Default(),
	}
	c.configuration.Monitor = s.URL
	c.configuration.Token = "1234"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Run(ctx)

	_, ok := h.getHost("foo")
	assert.False(t, ok)

	ch <- event{eventType: addEvent, ingress: &validIngress}
	assert.Eventually(t, func() bool {
		up, ok := h.getHost("example.com")
		return up && ok
	}, time.Second, time.Millisecond)

	ch <- event{eventType: deleteEvent, ingress: &validIngress}
	assert.Eventually(t, func() bool {
		up, ok := h.getHost("example.com")
		return !up && ok
	}, time.Second, time.Millisecond)

	s.Close()

	ch <- event{eventType: addEvent, ingress: &validIngress}
	assert.Never(t, func() bool {
		up, ok := h.getHost("foo")
		return up && ok
	}, time.Second, time.Millisecond)

}
