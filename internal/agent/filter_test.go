package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestFilter_Run(t *testing.T) {
	in := make(chan Event)
	out := make(chan Event, 1)
	f := filter{
		in:  in,
		out: out,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go f.Run(ctx)

	in <- Event{
		Host:        "foo",
		Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint},
	}

	ev := <-out
	assert.Equal(t, "foo", ev.Host)
}

func TestFilter_shouldForward(t *testing.T) {
	tests := []struct {
		name   string
		config Configuration
		event  Event
		want   assert.BoolAssertionFunc
	}{
		{
			name:   "pass",
			config: DefaultConfiguration,
			event:  Event{Type: AddEvent, Host: "foo", Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint}},
			want:   assert.True,
		},
		{
			name:   "no annotations",
			config: DefaultConfiguration,
			event:  Event{Type: AddEvent, Host: "foo"},
			want:   assert.False,
		},
		{
			name:   "no skip",
			config: Configuration{Hosts: map[string]EndpointConfiguration{"foo": DefaultGlobalConfiguration}},
			event:  Event{Type: AddEvent, Host: "foo", Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint}},
			want:   assert.True,
		},
		{
			name:   "skip",
			config: Configuration{Hosts: map[string]EndpointConfiguration{"foo": {Skip: true}}},
			event:  Event{Type: AddEvent, Host: "foo", Annotations: map[string]string{traefikEndpointAnnotation: traefikExternalEndpoint}},
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
