package monitor_test

import (
	"bytes"
	"github.com/clambin/uptime/internal/monitor"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMonitor(t *testing.T) {
	metrics := monitor.NewHTTPMetrics("uptime", "monitor")
	assert.NoError(t, testutil.CollectAndCompare(metrics, bytes.NewBufferString(``)))

	m := monitor.New(metrics, nil)

	req := monitor.Request{Target: "http://localhost", Interval: 10 * time.Millisecond}
	r, _ := http.NewRequest(http.MethodPost, "http://localhost:8080/target?"+req.Encode(), nil)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)

	assert.Eventually(t, func() bool {
		return testutil.CollectAndCount(metrics) > 0
	}, time.Second, 20*time.Millisecond)

	assert.NoError(t, testutil.CollectAndCompare(metrics, bytes.NewBufferString(`
# HELP uptime_monitor_up site is up/down
# TYPE uptime_monitor_up gauge
uptime_monitor_up{host="http://localhost"} 0
`), "uptime_monitor_up"))
}
