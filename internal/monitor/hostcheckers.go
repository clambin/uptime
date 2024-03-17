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

func (h *hostCheckers) add(target string, hostChecker checker, interval time.Duration) bool {
	h.lock.Lock()
	defer h.lock.Unlock()

	c, ok := h.hostCheckers[target]
	if ok {
		if c.GetRequest().Equals(hostChecker.GetRequest()) {
			return false
		}
		c.Cancel()
		delete(h.hostCheckers, target)
	}

	h.hostCheckers[target] = hostChecker
	go hostChecker.Run(interval)
	return true
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
