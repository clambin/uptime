package monitor

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _ slog.LogValuer = Request{}

type Request struct {
	Target    string
	Method    string
	ValidCode []int
	Interval  time.Duration
}

func (r Request) Encode() string {
	values := make(url.Values)
	values.Set("target", r.Target)
	if r.Method != "" {
		values.Set("method", r.Method)
	}
	if len(r.ValidCode) > 0 {
		codes := make([]string, len(r.ValidCode))
		for i := range r.ValidCode {
			codes[i] = strconv.Itoa(r.ValidCode[i])
		}
		values.Set("codes", strings.Join(codes, ","))
	}
	if r.Interval.Nanoseconds() > 0 {
		values.Set("interval", r.Interval.String())
	}
	return values.Encode()
}

func (r Request) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("target", r.Target),
		slog.String("method", r.Method),
		slog.Any("codes", r.ValidCode),
		slog.Duration("interval", r.Interval),
	)
}

func ParseRequest(r *http.Request) (Request, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return Request{}, fmt.Errorf("parse query: %w", err)
	}

	request := Request{
		Target: values.Get("target"),
		Method: values.Get("method"),
	}
	if request.Target == "" {
		return Request{}, errors.New("missing mandatory target")
	}
	if request.Method == "" {
		request.Method = http.MethodGet
	}

	codes := values.Get("codes")
	if codes == "" {
		codes = "200"
	}
	for _, codeString := range strings.Split(codes, ",") {
		code, err := strconv.Atoi(codeString)
		if err != nil {
			return Request{}, fmt.Errorf("invalid code %s: %w", codeString, err)
		}
		request.ValidCode = append(request.ValidCode, code)
	}
	interval := values.Get("interval")
	if interval == "" {
		interval = "5m"
	}
	if request.Interval, err = time.ParseDuration(interval); err != nil {
		return Request{}, fmt.Errorf("invalid interval %s: %w", interval, err)
	}
	return request, nil
}
