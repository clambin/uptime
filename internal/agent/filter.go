package agent

import "context"

type filter struct {
	in  <-chan Event
	out chan<- Event
}

func (f *filter) Run(ctx context.Context) {
	for {
		select {
		case ev := <-f.in:
			if f.shouldForward(ev) {
				f.out <- ev
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

func (f *filter) shouldForward(ev Event) bool {
	// TODO: make this configurable (how?)
	value, ok := ev.Annotations[traefikEndpointAnnotation]
	return ok && value == traefikExternalEndpoint
}
