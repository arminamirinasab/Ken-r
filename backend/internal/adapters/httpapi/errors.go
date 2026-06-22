package httpapi

import (
	"errors"
	"net/http"

	authdomain "github.com/kenar/backend/internal/domain/auth"
	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/domain/widget"
)

// writeError maps a domain error to an HTTP status and a localized message.
// Unexpected (5xx) errors are logged with the original cause; the client only
// ever sees a localized, non-leaking message.
func (a *api) writeError(w http.ResponseWriter, r *http.Request, err error) {
	status, key := mapError(err)
	if status >= http.StatusInternalServerError {
		a.log.Error("request failed", "path", r.URL.Path, "err", err)
	}
	loc := LocaleFrom(r.Context())
	writeJSON(w, status, map[string]string{"error": a.i18n.T(loc, key, nil)})
}

// mapError translates a domain error into (HTTP status, i18n key). The default
// is a 500 with a generic message so internal details never leak.
func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, authdomain.ErrInvalidPhone):
		return http.StatusBadRequest, "phone.invalid"
	case errors.Is(err, authdomain.ErrOTPNotFound):
		return http.StatusBadRequest, "otp.not_found"
	case errors.Is(err, authdomain.ErrOTPExpired):
		return http.StatusBadRequest, "otp.expired"
	case errors.Is(err, authdomain.ErrOTPInvalid):
		return http.StatusUnauthorized, "otp.invalid"
	case errors.Is(err, authdomain.ErrSessionInvalid):
		return http.StatusUnauthorized, "session.invalid"

	case errors.Is(err, pair.ErrAlreadyPaired):
		return http.StatusConflict, "pair.already_paired"
	case errors.Is(err, pair.ErrInviteNotFound):
		return http.StatusNotFound, "invite.invalid"
	case errors.Is(err, pair.ErrInviteExpired):
		return http.StatusGone, "invite.expired"
	case errors.Is(err, pair.ErrInviteUsed):
		return http.StatusConflict, "invite.used"
	case errors.Is(err, pair.ErrSelfPairing):
		return http.StatusBadRequest, "invite.self"
	case errors.Is(err, pair.ErrPairNotFound):
		return http.StatusNotFound, "pair.none"
	case errors.Is(err, pair.ErrNotMember):
		return http.StatusForbidden, "error.unauthorized"

	case errors.Is(err, widget.ErrUnknownKind):
		return http.StatusBadRequest, "widget.invalid"
	case errors.Is(err, widget.ErrEmptyPayload):
		return http.StatusBadRequest, "widget.payload_empty"
	case errors.Is(err, widget.ErrPayloadTooLarge):
		return http.StatusRequestEntityTooLarge, "widget.payload_too_large"

	default:
		return http.StatusInternalServerError, "error.internal"
	}
}

// badRequest writes a generic 400 for malformed request bodies.
func (a *api) badRequest(w http.ResponseWriter, r *http.Request) {
	loc := LocaleFrom(r.Context())
	writeJSON(w, http.StatusBadRequest, map[string]string{"error": a.i18n.T(loc, "error.bad_request", nil)})
}
