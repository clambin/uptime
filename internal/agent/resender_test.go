package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReSender_Run(t *testing.T) {
	in := make(chan event)
	out := make(chan event)
	r := reSender{
		in:     in,
		out:    out,
		events: make(map[string]event),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx, 500*time.Millisecond)

	evIn := event{eventType: addEvent, ingress: &validIngress}
	in <- evIn
	assert.Equal(t, evIn, <-out)
	assert.Equal(t, evIn, <-out)

	evIn.eventType = deleteEvent
	in <- evIn
	assert.Equal(t, evIn, <-out)
	assert.Never(t, func() bool {
		<-out
		return true
	}, 500*time.Millisecond, 100*time.Millisecond)
}
