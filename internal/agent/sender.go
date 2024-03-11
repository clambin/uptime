package agent

import (
	"context"
	"fmt"
	"github.com/clambin/uptime/internal/monitor"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Sender struct {
	Process    <-chan Event
	monitor    string
	token      string
	httpClient *http.Client
	logger     *slog.Logger
}

func (s Sender) Run(ctx context.Context) {
	for {
		select {
		case ev := <-s.Process:
			s.process(ctx, ev)
		case <-ctx.Done():
			return
		}
	}
}

func (s Sender) process(ctx context.Context, ev Event) {
	waitTime := time.Second
	l := s.logger.With("target", ev.Host)
	for {
		l.Debug("sending request")
		err := s.send(ctx, ev)
		if err == nil {
			return
		}
		l.Warn("request failed. waiting to retry", "err", err, "wait", waitTime)
		time.Sleep(waitTime)
		waitTime = min(waitTime*2, time.Minute)
	}
}

func (s Sender) makeRequest(ev Event) monitor.Request {
	// TODO: make this configurable
	return monitor.Request{
		Target:    ev.Host,
		Method:    http.MethodGet,
		ValidCode: []int{http.StatusOK, http.StatusForbidden},
		Interval:  5 * time.Minute,
	}
}

func (s Sender) send(ctx context.Context, ev Event) error {
	r, _ := http.NewRequestWithContext(ctx, getMethod(ev.Type), s.monitor+"?"+s.makeRequest(ev).Encode(), nil)
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
