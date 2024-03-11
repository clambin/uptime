package logger_test

import (
	"bytes"
	"github.com/clambin/uptime/pkg/logger"
	"github.com/clambin/uptime/pkg/logtester"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithLogger(t *testing.T) {
	var output bytes.Buffer
	l := logtester.New(&output, slog.LevelInfo)

	h := logger.WithLogger(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := r.Context().Value(logger.LoggerCtxValue).(*slog.Logger)
		l.Info("hello", "method", r.Method)
	}))

	r, _ := http.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "level=INFO msg=hello method=POST\n", output.String())
}
