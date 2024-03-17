package agent

import (
	"context"
	"time"
)

type reSender struct {
	in     <-chan event
	out    chan<- event
	events map[string]event
}

func (r *reSender) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case ev := <-r.in:
			eventKey := ev.namespace() + ":" + ev.name()
			switch ev.eventType {
			case addEvent:
				r.events[eventKey] = ev
			case deleteEvent:
				delete(r.events, eventKey)
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
