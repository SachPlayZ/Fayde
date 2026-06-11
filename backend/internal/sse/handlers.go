package sse

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SachPlayZ/rivz-asn/backend/internal/auth"
	"github.com/SachPlayZ/rivz-asn/backend/internal/httputil"
)

// Handler handles SSE connections.
type Handler struct {
	broker    *Broker
	jwtSecret string
}

// NewHandler creates a new SSE Handler.
func NewHandler(broker *Broker, jwtSecret string) *Handler {
	return &Handler{broker: broker, jwtSecret: jwtSecret}
}

// ServeSSE handles GET /events.
// The JWT token is read from the ?token= query parameter because EventSource
// cannot set custom headers.
func (h *Handler) ServeSSE(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		httputil.Error(w, http.StatusUnauthorized, "missing token")
		return
	}

	userID, role, err := auth.ValidateToken(tokenStr, h.jwtSecret)
	if err != nil {
		httputil.Error(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		httputil.Error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	flusher.Flush()

	var ch chan Event
	var unsub func()

	if role == "admin" {
		ch, unsub = h.broker.SubscribeAdmin()
	} else {
		ch, unsub = h.broker.Subscribe(userID)
	}
	defer unsub()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			b, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		}
	}
}
