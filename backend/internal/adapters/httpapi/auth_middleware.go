package httpapi

import (
	"context"
	"net/http"
	"strings"

	authdomain "github.com/kenar/backend/internal/domain/auth"
)

type userCtxKey int

const userKey userCtxKey = iota

// requireAuth wraps a handler, requiring a valid bearer session token. On
// success the resolved user id is stored in the request context (see userID).
func (a *api) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r.Header.Get("Authorization"))
		if token == "" {
			a.writeError(w, r, authdomain.ErrSessionInvalid)
			return
		}
		uid, err := a.auth.Authenticate(r.Context(), token)
		if err != nil {
			a.writeError(w, r, authdomain.ErrSessionInvalid)
			return
		}
		ctx := context.WithValue(r.Context(), userKey, uid)
		next(w, r.WithContext(ctx))
	}
}

// userID returns the authenticated user id stored by requireAuth.
func userID(ctx context.Context) string {
	if v, ok := ctx.Value(userKey).(string); ok {
		return v
	}
	return ""
}

// bearerToken extracts the token from an "Authorization: Bearer <token>" header.
func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}
