package monitor

import (
	"github.com/clambin/go-common/set"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestHostChecker_Up(t *testing.T) {
	tests := []struct {
		name     string
		response int
		valid    []int
		wantUp   assert.BoolAssertionFunc
	}{
		{
			name:     "up",
			response: http.StatusOK,
			valid:    []int{http.StatusOK},
			wantUp:   assert.True,
		},
		{
			name:     "down",
			response: http.StatusForbidden,
			valid:    []int{http.StatusOK},
			wantUp:   assert.False,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.response)
			}))
			defer s.Close()

			o := observer{}
			r := Request{
				Target:     strings.TrimPrefix(s.URL, "https://"),
				Method:     http.MethodGet,
				ValidCodes: set.New(tt.valid...),
			}
			h := newHostChecker(r, &o, s.Client(), slog.Default())
			assert.Equal(t, r, h.GetRequest())
			go h.Run(10 * time.Millisecond)

			var m HTTPMeasurement
			var ok bool
			assert.Eventually(t, func() bool {
				m, ok = o.result()
				return ok
			}, time.Second, 20*time.Millisecond)

			h.Cancel()

			tt.wantUp(t, m.Up)
			assert.True(t, m.IsTLS)
			assert.NotZero(t, m.TLSExpiry)
		})
	}
}

var _ HTTPObserver = &observer{}

type observer struct {
	observation HTTPMeasurement
	received    bool
	lock        sync.Mutex
}

func (o *observer) Observe(httpMetrics HTTPMeasurement) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.observation = httpMetrics
	o.received = true
}

func (o *observer) result() (HTTPMeasurement, bool) {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.observation, o.received
}
