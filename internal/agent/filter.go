package agent

import (
	"context"
	"log/slog"
)

type filter struct {
	in            <-chan Event
	out           chan<- Event
	configuration Configuration
	logger        *slog.Logger
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
	if !f.hasAnnotations(ev) {
		f.logger.Debug("host skipped: missing annotations", "host", ev.Host)
		return false
	}
	if !f.noSkip(ev) {
		f.logger.Debug("host skipped: on skip list", "host", ev.Host)
		return false
	}
	return true
}

func (f *filter) hasAnnotations(ev Event) bool {
	// TODO: make this configurable (how?)
	value, ok := ev.Annotations[traefikEndpointAnnotation]
	return ok && value == traefikExternalEndpoint
}

func (f *filter) noSkip(ev Event) bool {
	cfg, ok := f.configuration.Hosts[ev.Host]
	return !ok || !cfg.Skip
}
