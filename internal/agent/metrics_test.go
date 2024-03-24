package agent

import (
	"bytes"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics(t *testing.T) {
	m := NewMetrics("uptime", "agent", nil)

	ev := event{eventType: addEvent, ingress: &validIngress}
	m.ObserveEvent(ev)

	assert.NoError(t, testutil.CollectAndCompare(m, bytes.NewBufferString(`
# HELP uptime_agent_ingress_events_count number of ingress events received from kubernetes
# TYPE uptime_agent_ingress_events_count counter
uptime_agent_ingress_events_count{name="valid",namespace="foo",type="add"} 1
`)))
}
