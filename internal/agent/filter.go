package agent

import "context"

type filter struct {
	in            <-chan Event
	out           chan<- Event
	configuration Configuration
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
	return f.verifyHost(ev) && f.verifyAnnotations(ev)
}

func (f *filter) verifyHost(ev Event) bool {
	cfg, ok := f.configuration.Hosts[ev.Host]
	return !ok || !cfg.Skip
}

func (f *filter) verifyAnnotations(ev Event) bool {
	// TODO: make this configurable (how?)
	value, ok := ev.Annotations[traefikEndpointAnnotation]
	return ok && value == traefikExternalEndpoint
}
