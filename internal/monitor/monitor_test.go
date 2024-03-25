package monitor_test

import (
	"bytes"
	"github.com/clambin/uptime/internal/monitor"
	"github.com/clambin/uptime/internal/monitor/handlers"
	"github.com/clambin/uptime/internal/monitor/metrics"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitor(t *testing.T) {
	h := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	hm := metrics.NewHostMetrics("uptime", "monitor", nil)
	assert.NoError(t, testutil.CollectAndCompare(hm, bytes.NewBufferString(``)))

	mon := monitor.New(hm, http.DefaultClient)

	req := handlers.Request{Target: h.URL, Interval: 10 * time.Millisecond}
	r, _ := http.NewRequest(http.MethodPost, "/target?"+req.Encode(), nil)
	w := httptest.NewRecorder()

	mon.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	assert.Eventually(t, func() bool {
		return testutil.CollectAndCount(hm) > 0
	}, time.Second, 20*time.Millisecond)

	assert.NoError(t, testutil.CollectAndCompare(hm, bytes.NewBufferString(`
# HELP uptime_monitor_up site is up/down
# TYPE uptime_monitor_up gauge
uptime_monitor_up{host="`+h.URL+`"} 1
`), "uptime_monitor_up"))

	h.Close()

	assert.Eventually(t, func() bool {
		return nil == testutil.CollectAndCompare(hm, bytes.NewBufferString(`
# HELP uptime_monitor_up site is up/down
# TYPE uptime_monitor_up gauge
uptime_monitor_up{host="`+h.URL+`"} 0
`), "uptime_monitor_up")
	}, time.Second, 20*time.Millisecond)

	req = handlers.Request{Target: h.URL}
	r, _ = http.NewRequest(http.MethodDelete, "/target?"+req.Encode(), nil)
	w = httptest.NewRecorder()

	mon.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	// FIXME: deleted targets will continue to be reported on!
	assert.NoError(t, testutil.CollectAndCompare(hm, bytes.NewBufferString(`
# HELP uptime_monitor_up site is up/down
# TYPE uptime_monitor_up gauge
uptime_monitor_up{host="`+h.URL+`"} 0
`), "uptime_monitor_up"))
}
