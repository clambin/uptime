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
	in         <-chan Event
	monitor    string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
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
	// TODO: make this configurable
	return monitor.Request{
		Target:    ev.Host,
		Method:    http.MethodGet,
		ValidCode: []int{http.StatusOK, http.StatusUnauthorized},
		Interval:  5 * time.Minute,
	}
}

func (s sender) send(ctx context.Context, ev Event) error {
	r, _ := http.NewRequestWithContext(ctx, getMethod(ev.Type), s.monitor+"/target?"+s.makeRequest(ev).Encode(), nil)
	r.Header.Set("Authorization", "Bearer "+s.token)
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
