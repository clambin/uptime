package monitor

import (
	"time"
)

type HostCheckers map[string]*HostChecker

func (h HostCheckers) Add(target string, hostChecker *HostChecker, interval time.Duration) {
	_ = h.Remove(target)
	h[target] = hostChecker
	go h[target].Run(interval)
}

func (h HostCheckers) Remove(target string) bool {
	c, ok := h[target]
	if ok {
		c.Cancel()
		delete(h, target)
	}
	return ok
}
