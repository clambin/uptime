package handlers

import (
	"errors"
	"fmt"
	"github.com/clambin/go-common/set"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var _ slog.LogValuer = Request{}

type Request struct {
	Target     string
	Method     string
	ValidCodes set.Set[int]
	Interval   time.Duration
}

func (r Request) Equals(other Request) bool {
	return r.Target == other.Target &&
		r.Method == other.Method &&
		r.ValidCodes.Equals(other.ValidCodes) &&
		r.Interval == other.Interval
}

func (r Request) Encode() string {
	values := make(url.Values)
	values.Set("target", r.Target)
	if r.Method != "" {
		values.Set("method", r.Method)
	}
	if len(r.ValidCodes) > 0 {
		codes := make([]string, len(r.ValidCodes))
		for i, code := range r.ValidCodes.ListOrdered() {
			codes[i] = strconv.Itoa(code)
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
		slog.Any("codes", r.ValidCodes.ListOrdered()),
		slog.Duration("interval", r.Interval),
	)
}

func ParseRequest(r *http.Request) (Request, error) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return Request{}, fmt.Errorf("parse query: %w", err)
	}

	request := Request{
		Target:     values.Get("target"),
		Method:     values.Get("method"),
		ValidCodes: set.New[int](),
	}
	if request.Target == "" {
		return Request{}, errors.New("missing mandatory target")
	}
	if request.Method == "" {
		request.Method = http.MethodGet
	}

	if codes := values.Get("codes"); codes != "" {
		for _, code := range strings.Split(codes, ",") {
			val, err := strconv.Atoi(code)
			if err != nil {
				return Request{}, fmt.Errorf("invalid code %s: %w", code, err)
			}
			request.ValidCodes.Add(val)
		}
	}
	if len(request.ValidCodes) == 0 {
		request.ValidCodes.Add(http.StatusOK)
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
