package retry

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMultiplyingWaiter(t *testing.T) {
	w := MultiplyingWaiter{InitialWait: time.Millisecond, MaxWait: 100 * time.Millisecond, Factor: 2}
	for range 7 {
		assert.NoError(t, w.Wait(context.Background()))
	}
	assert.Equal(t, 100*time.Millisecond, w.currentWait)

	w.Reset()
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	assert.Eventually(t, func() bool {
		err := w.Wait(ctx)
		return errors.Is(err, ErrInterrupted)
	}, time.Second, 100*time.Millisecond)
}
