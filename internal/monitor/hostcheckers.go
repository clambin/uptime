package monitor

import (
	"sync"
	"time"
)

type checker interface {
	Run(time.Duration)
	Cancel()
	GetRequest() Request
}

type hostCheckers struct {
	hostCheckers map[string]checker
	lock         sync.Mutex
}

func (h *hostCheckers) add(target string, hostChecker checker, interval time.Duration) {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[target]
	if ok {
		if c.GetRequest().Equals(hostChecker.GetRequest()) {
			return
		}
		c.Cancel()
		delete(h.hostCheckers, target)
	}

	h.hostCheckers[target] = hostChecker
	go hostChecker.Run(interval)
}

func (h *hostCheckers) remove(target string) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[target]
	if ok {
		c.Cancel()
		delete(h.hostCheckers, target)
	}
	return ok
}
