package monitor

import (
	"sync"
	"time"
)

type HostCheckers struct {
	hostCheckers map[string]*HostChecker
	lock         sync.Mutex
}

func (h *HostCheckers) Add(target string, hostChecker *HostChecker, interval time.Duration) {
	_ = h.Remove(target)

	h.lock.Lock()
	defer h.lock.Unlock()

	h.hostCheckers[target] = hostChecker
	go hostChecker.Run(interval)
}

func (h *HostCheckers) Remove(target string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[target]
	if ok {
		c.Cancel()
		delete(h.hostCheckers, target)
	}
	return ok
}
