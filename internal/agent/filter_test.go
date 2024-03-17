package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestFilter_Run(t *testing.T) {
	in := make(chan event)
	out := make(chan event, 1)
	f := filter{
		in:  in,
		out: out,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go f.Run(ctx)

	evIn := event{eventType: addEvent, ingress: &validIngress}
	in <- evIn
	assert.Equal(t, evIn, <-out)
}

func TestFilter_shouldForward(t *testing.T) {
	tests := []struct {
		name   string
		config Configuration
		event  event
		want   assert.BoolAssertionFunc
	}{
		{
			name:   "pass",
			config: DefaultConfiguration,
			event:  event{eventType: addEvent, ingress: &validIngress},
			want:   assert.True,
		},
		{
			name:   "no annotations",
			config: DefaultConfiguration,
			event:  event{eventType: addEvent, ingress: &invalidIngress},
			want:   assert.False,
		},
		{
			name:   "no skip",
			config: Configuration{Hosts: map[string]EndpointConfiguration{"foo.com": DefaultGlobalConfiguration}},
			event:  event{eventType: addEvent, ingress: &validIngress},
			want:   assert.True,
		},
		{
			name:   "skip",
			config: Configuration{Hosts: map[string]EndpointConfiguration{"example.com": {Skip: true}}},
			event:  event{eventType: addEvent, ingress: &validIngress},
			want:   assert.False,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f := filter{configuration: tt.config, logger: slog.Default()}
			tt.want(t, f.shouldForward(tt.event))
		})
	}
}
