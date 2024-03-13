package retry

import (
	"context"
	"errors"
	"time"
)

type Waiter interface {
	Wait(ctx context.Context) error
	Reset()
}

var ErrInterrupted = errors.New("wait interrupted")

var _ Waiter = &MultiplyingWaiter{}

type MultiplyingWaiter struct {
	InitialWait time.Duration
	MaxWait     time.Duration
	Factor      float64
	currentWait time.Duration
}

func (w *MultiplyingWaiter) Wait(ctx context.Context) error {
	if w.currentWait == 0 {
		w.currentWait = w.InitialWait
	}
	select {
	case <-time.After(w.currentWait):
		w.currentWait = min(w.MaxWait, time.Duration(w.Factor*float64(w.currentWait)))
		return nil
	case <-ctx.Done():
		return ErrInterrupted
	}
}

func (w *MultiplyingWaiter) Reset() {
	w.currentWait = w.InitialWait
}
