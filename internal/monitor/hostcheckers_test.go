package monitor

import (
	"github.com/clambin/go-common/set"
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestHostCheckers(t *testing.T) {
	const target = "example.com"
	req := handlers.Request{
		Target:     target,
		Method:     http.MethodGet,
		ValidCodes: set.New(http.StatusOK),
		Interval:   time.Minute,
	}
	l := slog.Default()

	checkers := hostCheckers{
		metrics:      NewHostMetrics("", "", nil),
		hostCheckers: make(map[string]*hostChecker),
	}

	checkers.Add(req, l)
	p, ok := checkers.hostCheckers[req.Target]
	assert.True(t, ok)

	checkers.Add(req, l)
	p2, ok := checkers.hostCheckers[req.Target]
	assert.True(t, ok)
	assert.Equal(t, p, p2)

	req.Interval = time.Hour
	checkers.Add(req, l)
	p2, ok = checkers.hostCheckers[req.Target]
	assert.True(t, ok)
	assert.NotEqual(t, p, p2)

	checkers.Remove(req)
	_, ok = checkers.hostCheckers[req.Target]
	assert.False(t, ok)
}
