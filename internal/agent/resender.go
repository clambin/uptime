package agent

import (
	"context"
	"time"
)

type reSender struct {
	in     <-chan Event
	out    chan<- Event
	events map[string]Event
}

func (r *reSender) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case ev := <-r.in:
			switch ev.Type {
			case AddEvent:
				r.events[ev.Host] = ev
			case DeleteEvent:
				delete(r.events, ev.Host)
			}
			r.out <- ev
		case <-ticker.C:
			for _, ev := range r.events {
				r.out <- ev
			}
		case <-ctx.Done():
			return
		}
	}
}
