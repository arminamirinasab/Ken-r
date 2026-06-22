package httpapi

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/kenar/backend/internal/domain/widget"
)

// localized returns the localized text for key in the request's locale.
func (a *api) localized(r *http.Request, key string) string {
	return a.i18n.T(LocaleFrom(r.Context()), key, nil)
}

// --- Auth ---

func (a *api) handleOTPRequest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone string `json:"phone"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		a.badRequest(w, r)
		return
	}
	phone, err := a.auth.RequestOTP(r.Context(), req.Phone, LocaleFrom(r.Context()))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"phone":   phone,
		"message": a.localized(r, "otp.sent"),
	})
}

func (a *api) handleOTPVerify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Phone       string `json:"phone"`
		Code        string `json:"code"`
		DisplayName string `json:"display_name"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		a.badRequest(w, r)
		return
	}
	token, u, err := a.auth.VerifyOTP(r.Context(), req.Phone, req.Code, req.DisplayName)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"token":        token,
		"user_id":      u.ID,
		"display_name": u.DisplayName,
	})
}

// --- Pairing ---

func (a *api) handleCreateInvite(w http.ResponseWriter, r *http.Request) {
	inv, err := a.pairing.CreateInvite(r.Context(), userID(r.Context()))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"code":       inv.Code,
		"expires_at": inv.ExpiresAt.UTC(),
		"message":    a.localized(r, "invite.created"),
	})
}

func (a *api) handleAcceptInvite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		a.badRequest(w, r)
		return
	}
	p, err := a.pairing.AcceptInvite(r.Context(), req.Code, userID(r.Context()))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"pair_id": p.ID,
		"message": a.localized(r, "invite.accepted"),
	})
}

func (a *api) handleSpace(w http.ResponseWriter, r *http.Request) {
	uid := userID(r.Context())
	p, err := a.pairing.Space(r.Context(), uid)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	partnerID, _ := p.Partner(uid)
	writeJSON(w, http.StatusOK, map[string]any{
		"pair_id":    p.ID,
		"partner_id": partnerID,
		"status":     string(p.Status),
		"premium":    p.IsPremium(time.Now()),
		"created_at": p.CreatedAt.UTC(),
	})
}

func (a *api) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if err := a.pairing.Disconnect(r.Context(), userID(r.Context())); err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": a.localized(r, "pair.disconnected")})
}

// --- Devices ---

func (a *api) handleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PushToken string `json:"push_token"`
		Provider  string `json:"provider"`
	}
	if err := decodeJSON(w, r, &req); err != nil || req.PushToken == "" {
		a.badRequest(w, r)
		return
	}
	if req.Provider == "" {
		req.Provider = "pushe"
	}
	if err := a.devices.Upsert(r.Context(), userID(r.Context()), req.PushToken, req.Provider); err != nil {
		a.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Widgets ---

func (a *api) handleSetWidget(w http.ResponseWriter, r *http.Request) {
	kind := widget.Kind(r.PathValue("kind"))
	var req struct {
		Payload string            `json:"payload"` // base64-encoded E2E blob
		Meta    map[string]string `json:"meta"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		a.badRequest(w, r)
		return
	}
	payload, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		a.badRequest(w, r)
		return
	}
	st, err := a.widgets.Set(r.Context(), userID(r.Context()), kind, payload, req.Meta)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": st.Version,
		"message": a.localized(r, "widget.updated"),
	})
}

func (a *api) handleGetWidget(w http.ResponseWriter, r *http.Request) {
	kind := widget.Kind(r.PathValue("kind"))
	states, err := a.widgets.Latest(r.Context(), userID(r.Context()), kind)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(states))
	for _, st := range states {
		out = append(out, map[string]any{
			"author_id":    st.AuthorID,
			"payload":      base64.StdEncoding.EncodeToString(st.Payload),
			"payload_meta": st.PayloadMeta,
			"version":      st.Version,
			"updated_at":   st.UpdatedAt.UTC(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"states": out})
}
