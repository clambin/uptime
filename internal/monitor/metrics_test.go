package monitor

import (
	"bytes"
	"github.com/clambin/uptime/pkg/logtester"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

func TestHTTPMetrics_Observe(t *testing.T) {
	metrics := NewHTTPMetrics()
	assert.NoError(t, testutil.CollectAndCompare(metrics, bytes.NewBufferString(``)))

	metrics.Observe(HTTPMeasurement{
		Host:      "localhost",
		Up:        true,
		Code:      http.StatusOK,
		Latency:   time.Second,
		IsTLS:     true,
		TLSExpiry: time.Hour,
	})

	assert.NoError(t, testutil.CollectAndCompare(metrics, bytes.NewBufferString(`
# HELP uptime_monitor_certificate_expiry_days number of days before the certificate expires
# TYPE uptime_monitor_certificate_expiry_days gauge
uptime_monitor_certificate_expiry_days{host="localhost"} 0.041666666666666664
# HELP uptime_monitor_up site is up/down
# TYPE uptime_monitor_up gauge
uptime_monitor_up{host="localhost"} 1
`)))
}

func TestHTTPMeasurement_LogValue(t *testing.T) {
	tests := []struct {
		name string
		m    HTTPMeasurement
		want string
	}{
		{
			name: "down",
			m:    HTTPMeasurement{Host: "localhost"},
			want: "level=INFO msg=measurement m.target=localhost m.up=false\n",
		},
		{
			name: "rejected",
			m:    HTTPMeasurement{Host: "localhost", Code: http.StatusInternalServerError, Latency: time.Millisecond},
			want: "level=INFO msg=measurement m.target=localhost m.up=false m.code=500 m.latency=1ms\n",
		},
		{
			name: "up",
			m:    HTTPMeasurement{Host: "localhost", Up: true, Code: http.StatusOK, Latency: time.Millisecond, IsTLS: true, TLSExpiry: time.Hour},
			want: "level=INFO msg=measurement m.target=localhost m.up=true m.code=200 m.latency=1ms m.certExpiry=1h0m0s\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer
			l := logtester.New(&output, slog.LevelInfo)
			l.Info("measurement", "m", tt.m)
			assert.Equal(t, tt.want, output.String())
		})
	}
}
