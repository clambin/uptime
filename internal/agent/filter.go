package agent

import "context"

type Filter struct {
	EventsIn  <-chan Event
	EventsOut chan<- Event
}

func (f *Filter) Run(ctx context.Context) {
	for {
		select {
		case ev := <-f.EventsIn:
			if f.shouldForward(ev) {
				f.EventsOut <- ev
			}
		case <-ctx.Done():
			return
		}
	}
}

const (
	traefikEndpointAnnotation = "traefik.ingress.kubernetes.io/router.entrypoints"
	traefikExternalEndpoint   = "websecure"
)

func (f *Filter) shouldForward(ev Event) bool {
	// TODO: make this configurable (how?)
	value, ok := ev.Annotations[traefikEndpointAnnotation]
	return ok && value == traefikExternalEndpoint
}
