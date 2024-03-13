package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
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
