package logtester

import (
	"io"
	"log/slog"
)

func New(output io.Writer, level slog.Level) *slog.Logger {
	opts := slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	}
	return slog.New(slog.NewTextHandler(output, &opts))
}
