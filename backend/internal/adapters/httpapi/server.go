// Package httpapi is the HTTP transport adapter. It wires routes to use cases
// and translates domain results/errors into JSON responses with localized text.
package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/kenar/backend/internal/app/auth"
	"github.com/kenar/backend/internal/app/pairing"
	"github.com/kenar/backend/internal/app/widgets"
	"github.com/kenar/backend/internal/platform/i18n"
	"github.com/kenar/backend/internal/ports"
)

// Deps are the collaborators the HTTP layer needs.
type Deps struct {
	Logger  *slog.Logger
	I18n    *i18n.Bundle
	Auth    *auth.Service
	Pairing *pairing.Service
	Widgets *widgets.Service
	Devices ports.DeviceRepo
	Events  ports.EventSubscriber
}

// api bundles the dependencies so handlers can be methods.
type api struct {
	log     *slog.Logger
	i18n    *i18n.Bundle
	auth    *auth.Service
	pairing *pairing.Service
	widgets *widgets.Service
	devices ports.DeviceRepo
	events  ports.EventSubscriber
}

// NewRouter builds the HTTP handler with middleware applied.
func NewRouter(d Deps) http.Handler {
	a := &api{
		log:     d.Logger,
		i18n:    d.I18n,
		auth:    d.Auth,
		pairing: d.Pairing,
		widgets: d.Widgets,
		devices: d.Devices,
		events:  d.Events,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth (public)
	mux.HandleFunc("POST /v1/auth/otp/request", a.handleOTPRequest)
	mux.HandleFunc("POST /v1/auth/otp/verify", a.handleOTPVerify)

	// Pairing (authenticated)
	mux.HandleFunc("POST /v1/invites", a.requireAuth(a.handleCreateInvite))
	mux.HandleFunc("POST /v1/invites/accept", a.requireAuth(a.handleAcceptInvite))
	mux.HandleFunc("GET /v1/space", a.requireAuth(a.handleSpace))
	mux.HandleFunc("POST /v1/space/disconnect", a.requireAuth(a.handleDisconnect))

	// Devices (authenticated)
	mux.HandleFunc("POST /v1/devices", a.requireAuth(a.handleRegisterDevice))

	// Widgets (authenticated)
	mux.HandleFunc("POST /v1/widgets/{kind}", a.requireAuth(a.handleSetWidget))
	mux.HandleFunc("GET /v1/widgets/{kind}", a.requireAuth(a.handleGetWidget))

	// Real-time delivery stream (authenticated). SSE is the stdlib interim
	// transport; a WebSocket adapter (same event bus) lands for bidirectional
	// presence features. See ROADMAP §"Real-time backbone".
	mux.HandleFunc("GET /v1/stream", a.requireAuth(a.handleStream))

	return LocaleMiddleware(d.I18n)(logRequests(d.Logger)(mux))
}

func logRequests(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info("request", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// decodeJSON reads a JSON request body (size-limited) into dst.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MiB ceiling
	return json.NewDecoder(r.Body).Decode(dst)
}
