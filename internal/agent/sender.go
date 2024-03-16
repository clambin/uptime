package agent

import (
	"context"
	"fmt"
	"github.com/clambin/uptime/internal/monitor"
	"github.com/clambin/uptime/pkg/retry"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type sender struct {
	in            <-chan event
	configuration Configuration
	metrics       *Metrics
	httpClient    *http.Client
	logger        *slog.Logger
}

func (s sender) Run(ctx context.Context) {
	for {
		select {
		case ev := <-s.in:
			s.process(ctx, ev)
		case <-ctx.Done():
			return
		}
	}
}

func (s sender) process(ctx context.Context, ev event) {
	l := s.logger.With("event", ev)
	l.Debug("sending request")

	method := getMethod(ev.eventType)
	for _, request := range s.makeRequests(ev) {
		waiter := retry.MultiplyingWaiter{InitialWait: time.Second, MaxWait: time.Millisecond, Factor: 2}
		for {
			err := s.send(ctx, method, request)
			if err == nil {
				return
			}
			l.Warn("request failed. waiting to retry", "err", err)
			if waiter.Wait(ctx) != nil {
				return
			}
		}
	}
}

func getMethod(ev eventType) string {
	switch ev {
	case addEvent:
		return http.MethodPost
	case deleteEvent:
		return http.MethodDelete
	default:
		panic("invalid event type: " + ev)
	}
}

func (s sender) makeRequests(ev event) []monitor.Request {
	targets := ev.targetHosts()
	requests := make([]monitor.Request, len(targets))
	for i := range targets {
		requests[i] = s.makeRequest(targets[i])
	}
	return requests
}

func (s sender) makeRequest(host string) monitor.Request {
	ep := s.configuration.Global
	if custom, ok := s.configuration.Hosts[host]; ok {
		if custom.Method != "" {
			ep.Method = custom.Method
		}
		if custom.Interval != 0 {
			ep.Interval = custom.Interval
		}
		if custom.ValidStatusCodes != nil {
			ep.ValidStatusCodes = custom.ValidStatusCodes
		}
	}
	return monitor.Request{
		// TODO: best place to manage adding https:// ???
		Target:    "https://" + host,
		Method:    ep.Method,
		ValidCode: ep.ValidStatusCodes,
		Interval:  ep.Interval,
	}
}

func (s sender) send(ctx context.Context, method string, request monitor.Request) error {
	start := time.Now()
	r, _ := http.NewRequestWithContext(ctx, method, s.configuration.Monitor+"/target?"+request.Encode(), nil)
	if s.configuration.Token != "" {
		r.Header.Set("Authorization", "Bearer "+s.configuration.Token)
	}
	resp, err := s.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	if s.metrics != nil {
		s.metrics.ObserveRequest(resp.StatusCode, time.Since(start))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status: %s", resp.Status)
	}

	return nil
}
