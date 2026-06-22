package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// streamHeartbeat keeps the SSE connection (and intermediary proxies) alive
// between events.
const streamHeartbeat = 25 * time.Second

// handleStream delivers the caller's pair events as Server-Sent Events. The
// client subscribes after authenticating; whenever either partner updates a
// widget, an event is pushed here so the app can refresh its Glance widget.
//
// SSE is the stdlib interim transport for the real-time backbone; a WebSocket
// adapter (sharing the same event bus) replaces/augments it for bidirectional
// presence features (Presence Pulse, Hold Hands).
func (a *api) handleStream(w http.ResponseWriter, r *http.Request) {
	uid := userID(r.Context())
	p, err := a.pairing.Space(r.Context(), uid)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		a.writeError(w, r, fmt.Errorf("streaming unsupported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	ctx := r.Context()
	events, err := a.events.Subscribe(ctx, p.ID)
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	ticker := time.NewTicker(streamHeartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case ev, open := <-events:
			if !open {
				return
			}
			data, err := json.Marshal(ev)
			if err != nil {
				a.log.Error("stream marshal", "err", err)
				continue
			}
			fmt.Fprintf(w, "event: widget\ndata: %s\n\n", data)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}
