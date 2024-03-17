package agent

import (
	"context"
	"log/slog"
)

type filter struct {
	in            <-chan event
	out           chan<- event
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

func (f *filter) shouldForward(ev event) bool {
	if !f.hasAnnotations(ev) {
		f.logger.Debug("ingress skipped: missing annotations", "event", ev)
		return false
	}
	if f.skip(ev) {
		f.logger.Debug("ingress skipped: host on skip list", "event", ev)
		return false
	}
	return true
}

func (f *filter) hasAnnotations(ev event) bool {
	// TODO: make this configurable
	return ev.hasAnnotation(traefikEndpointAnnotation, traefikExternalEndpoint)
}

func (f *filter) skip(ev event) bool {
	// TODO: if any of the hosts are on the skip list, we skip the entire ingress. make it more granular?
	for _, host := range ev.targetHosts() {
		cfg, ok := f.configuration.Hosts[host]
		if ok && cfg.Skip {
			return true
		}
	}
	return false
}
