package agent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestReSender_Run(t *testing.T) {
	in := make(chan Event)
	out := make(chan Event)
	r := reSender{
		in:     in,
		out:    out,
		events: make(map[string]Event),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go r.Run(ctx, 500*time.Millisecond)

	in <- Event{Type: AddEvent, Host: "foo"}
	assert.Equal(t, Event{Type: AddEvent, Host: "foo"}, <-out)
	assert.Equal(t, Event{Type: AddEvent, Host: "foo"}, <-out)

	in <- Event{Type: DeleteEvent, Host: "foo"}
	assert.Equal(t, Event{Type: DeleteEvent, Host: "foo"}, <-out)
	assert.Never(t, func() bool {
		<-out
		return true
	}, 500*time.Millisecond, 100*time.Millisecond)

}
