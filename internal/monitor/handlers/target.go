package handlers

import (
	"github.com/clambin/uptime/pkg/logger"
	"log/slog"
	"net/http"
)

var _ http.Handler = &TargetHandler{}

type TargetHandler struct {
	TargetManager
}

type TargetManager interface {
	Add(Request, *slog.Logger)
	Remove(Request)
}

func (t TargetHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	l := logger.Logger(req)
	r, err := ParseRequest(req)
	if err != nil {
		l.Error("invalid request", "err", err)
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	switch req.Method {
	case http.MethodPost:
		t.Add(r, l)
	case http.MethodDelete:
		t.Remove(r)
	default:
		http.Error(w, "invalid method: "+req.Method, http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
}
