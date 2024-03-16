package agent

import (
	"bytes"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	m := NewMetrics("uptime", "agent")

	ev := event{eventType: addEvent, ingress: &validIngress}
	m.ObserveEvent(ev)
	m.ObserveRequest(http.StatusOK, time.Second)

	assert.NoError(t, testutil.CollectAndCompare(m, bytes.NewBufferString(`
# HELP uptime_agent_ingress_events_count number of ingress events received from kubernetes
# TYPE uptime_agent_ingress_events_count counter
uptime_agent_ingress_events_count{name="valid",namespace="foo",type="add"} 1
# HELP uptime_agent_request_latency latency of requests sent to the monitor
# TYPE uptime_agent_request_latency histogram
uptime_agent_request_latency_bucket{code="200",le="0.01"} 0
uptime_agent_request_latency_bucket{code="200",le="0.1"} 0
uptime_agent_request_latency_bucket{code="200",le="0.2"} 0
uptime_agent_request_latency_bucket{code="200",le="0.5"} 0
uptime_agent_request_latency_bucket{code="200",le="1"} 1
uptime_agent_request_latency_bucket{code="200",le="2"} 1
uptime_agent_request_latency_bucket{code="200",le="+Inf"} 1
uptime_agent_request_latency_sum{code="200"} 1
uptime_agent_request_latency_count{code="200"} 1
# HELP uptime_agent_requests_count number of requests sent to the monitor
# TYPE uptime_agent_requests_count counter
uptime_agent_requests_count{code="200"} 1
`)))

}
