package monitor

import (
	"github.com/clambin/go-common/set"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestHostCheckers(t *testing.T) {
	const target = "example.com"
	req := Request{
		Target:     target,
		Method:     http.MethodGet,
		ValidCodes: set.New(http.StatusOK),
		Interval:   time.Minute,
	}

	checkers := hostCheckers{
		hostCheckers: make(map[string]checker),
	}

	c1 := fakeChecker{req: req}
	ok := checkers.add(target, &c1, time.Millisecond)
	assert.True(t, ok)

	assert.Eventually(t, func() bool {
		return c1.running.Load()
	}, time.Second, time.Millisecond)

	req.Interval = time.Hour
	c2 := fakeChecker{req: req}

	ok = checkers.add(target, &c2, time.Millisecond)
	assert.True(t, ok)

	assert.Eventually(t, func() bool {
		return c2.running.Load()
	}, time.Second, time.Millisecond)
	assert.False(t, c1.running.Load())

	c3 := fakeChecker{req: req}
	ok = checkers.add(target, &c3, time.Millisecond)
	assert.False(t, ok)

	checkers.remove(target)
	assert.Eventually(t, func() bool {
		return !c1.running.Load()
	}, time.Second, time.Millisecond)
}

var _ checker = &fakeChecker{}

type fakeChecker struct {
	running atomic.Bool
	req     Request
}

func (f *fakeChecker) Run(_ time.Duration) {
	f.running.Store(true)
}

func (f *fakeChecker) Cancel() {
	f.running.Store(false)
}

func (f *fakeChecker) GetRequest() Request {
	return f.req
}
