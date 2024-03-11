package monitor

import (
	"bytes"
	"github.com/clambin/uptime/pkg/logtester"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestHTTPMeasurement_LogValue(t *testing.T) {
	tests := []struct {
		name string
		m    HTTPMeasurement
		want string
	}{
		{
			name: "down",
			m:    HTTPMeasurement{Host: "http://localhost"},
			want: "level=INFO msg=measurement m.target=http://localhost m.up=false\n",
		},
		{
			name: "rejected",
			m:    HTTPMeasurement{Host: "http://localhost", Code: "500", Latency: time.Millisecond},
			want: "level=INFO msg=measurement m.target=http://localhost m.up=false m.code=500 m.latency=1ms\n",
		},
		{
			name: "up",
			m:    HTTPMeasurement{Host: "http://localhost", Up: true, Code: "200", Latency: time.Millisecond, IsTLS: true, TLSExpiry: time.Hour},
			want: "level=INFO msg=measurement m.target=http://localhost m.up=true m.code=200 m.latency=1ms m.certExpiry=1h0m0s\n",
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
