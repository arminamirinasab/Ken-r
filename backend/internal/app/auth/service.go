// Package auth implements phone-OTP authentication use cases: requesting a
// login code (sent via the SMSProvider port) and verifying it to mint a session
// token, creating the user on first login. It depends only on ports and the
// auth domain — never on concrete adapters.
package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	authdomain "github.com/kenar/backend/internal/domain/auth"
	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/platform/i18n"
	"github.com/kenar/backend/internal/ports"
)

// OTPTTL is how long a requested login code stays valid.
const OTPTTL = 5 * time.Minute

// SessionTTL is how long an issued session token stays valid.
const SessionTTL = 30 * 24 * time.Hour

// Clock lets tests control time; production uses time.Now.
type Clock func() time.Time

// Service holds the authentication use cases.
type Service struct {
	users    ports.UserRepo
	otps     ports.OTPRepo
	sessions ports.SessionRepo
	sms      ports.SMSProvider
	i18n     *i18n.Bundle
	now      Clock
}

// New constructs an auth Service.
func New(users ports.UserRepo, otps ports.OTPRepo, sessions ports.SessionRepo, sms ports.SMSProvider, bundle *i18n.Bundle, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{users: users, otps: otps, sessions: sessions, sms: sms, i18n: bundle, now: clock}
}

// RequestOTP normalizes the phone, generates a login code, stores it, and sends
// it via SMS in the requested locale. The canonical phone is returned so the
// caller can echo it back for the verify step.
func (s *Service) RequestOTP(ctx context.Context, rawPhone, locale string) (string, error) {
	phone, err := authdomain.NormalizePhone(rawPhone)
	if err != nil {
		return "", err
	}
	code, err := authdomain.NewOTP()
	if err != nil {
		return "", fmt.Errorf("auth: generate otp: %w", err)
	}
	if err := s.otps.Put(ctx, phone, code, s.now().Add(OTPTTL)); err != nil {
		return "", fmt.Errorf("auth: store otp: %w", err)
	}
	body := s.i18n.T(locale, "otp.sms.body", map[string]string{"code": code})
	if err := s.sms.Send(ctx, phone, body); err != nil {
		return "", fmt.Errorf("auth: send otp: %w", err)
	}
	return phone, nil
}

// VerifyOTP checks the submitted code for the phone and, on success, returns a
// new session token and the user (created on first login). The code is consumed
// whether or not it matched only on success — a wrong code can be retried until
// it expires.
func (s *Service) VerifyOTP(ctx context.Context, rawPhone, code, displayName string) (string, pair.User, error) {
	phone, err := authdomain.NormalizePhone(rawPhone)
	if err != nil {
		return "", pair.User{}, err
	}
	stored, expiresAt, err := s.otps.Get(ctx, phone)
	if err != nil {
		return "", pair.User{}, err // authdomain.ErrOTPNotFound
	}
	if !s.now().Before(expiresAt) {
		_ = s.otps.Delete(ctx, phone)
		return "", pair.User{}, authdomain.ErrOTPExpired
	}
	if subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		return "", pair.User{}, authdomain.ErrOTPInvalid
	}
	_ = s.otps.Delete(ctx, phone)

	u, err := s.upsertUser(ctx, phone, displayName)
	if err != nil {
		return "", pair.User{}, err
	}
	token, err := s.sessions.Create(ctx, u.ID, s.now().Add(SessionTTL))
	if err != nil {
		return "", pair.User{}, fmt.Errorf("auth: create session: %w", err)
	}
	return token, u, nil
}

// Authenticate resolves a bearer session token to its user id.
func (s *Service) Authenticate(ctx context.Context, token string) (string, error) {
	return s.sessions.UserID(ctx, token)
}

func (s *Service) upsertUser(ctx context.Context, phone, displayName string) (pair.User, error) {
	u, err := s.users.GetByPhone(ctx, phone)
	if err == nil {
		return u, nil
	}
	if !errors.Is(err, pair.ErrUserNotFound) {
		return pair.User{}, fmt.Errorf("auth: lookup user: %w", err)
	}
	created, err := s.users.Create(ctx, pair.User{
		Phone:       phone,
		DisplayName: displayName,
		Locale:      pair.LocaleFa,
		CreatedAt:   s.now(),
	})
	if err != nil {
		return pair.User{}, fmt.Errorf("auth: create user: %w", err)
	}
	return created, nil
}
