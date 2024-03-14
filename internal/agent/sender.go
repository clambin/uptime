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
	in            <-chan Event
	configuration Configuration
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

func (s sender) process(ctx context.Context, ev Event) {
	l := s.logger.With("target", ev.Host, "eventType", ev.Type)
	l.Debug("sending request")
	waiter := retry.MultiplyingWaiter{InitialWait: time.Second, MaxWait: time.Millisecond, Factor: 2}
	for {
		err := s.send(ctx, ev)
		if err == nil {
			return
		}
		l.Warn("request failed. waiting to retry", "err", err)
		if waiter.Wait(ctx) != nil {
			return
		}
	}
}

func (s sender) makeRequest(ev Event) monitor.Request {
	ep, ok := s.configuration.Hosts[ev.Host]
	if !ok {
		ep = s.configuration.Global
	}
	return monitor.Request{
		Target:    ev.Host,
		Method:    ep.Method,
		ValidCode: ep.ValidStatusCodes,
		Interval:  ep.Interval,
	}
}

func (s sender) send(ctx context.Context, ev Event) error {
	r, _ := http.NewRequestWithContext(ctx, getMethod(ev.Type), s.configuration.Monitor+"/target?"+s.makeRequest(ev).Encode(), nil)
	if s.configuration.Token != "" {
		r.Header.Set("Authorization", "Bearer "+s.configuration.Token)
	}
	resp, err := s.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status: %s", resp.Status)
	}

	return nil
}

func getMethod(ev EventType) string {
	switch ev {
	case AddEvent:
		return http.MethodPost
	case DeleteEvent:
		return http.MethodDelete
	default:
		panic("invalid event type: " + ev)
	}
}
