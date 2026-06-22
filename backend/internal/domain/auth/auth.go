// Package auth holds the core domain for phone-OTP authentication: domain
// errors, OTP code generation, and Iranian phone-number normalization. No
// infrastructure dependencies.
package auth

import (
	"crypto/rand"
	"errors"
	"strings"
)

// Domain errors. Adapters map these to transport-level responses + i18n keys.
var (
	ErrInvalidPhone   = errors.New("invalid phone number")
	ErrOTPNotFound    = errors.New("no login code was requested for this phone")
	ErrOTPExpired     = errors.New("login code expired")
	ErrOTPInvalid     = errors.New("login code is incorrect")
	ErrSessionInvalid = errors.New("session is invalid or expired")
)

// OTPDigits is the length of a generated login code.
const OTPDigits = 6

// NewOTP returns a cryptographically random numeric login code of OTPDigits
// length (zero-padded, leading zeros allowed).
func NewOTP() (string, error) {
	const digits = "0123456789"
	buf := make([]byte, OTPDigits)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	var sb strings.Builder
	sb.Grow(OTPDigits)
	for _, b := range buf {
		sb.WriteByte(digits[int(b)%len(digits)])
	}
	return sb.String(), nil
}

// NormalizePhone validates an Iranian mobile number and canonicalizes it to
// E.164 (+989XXXXXXXXX). It accepts the common local forms:
//
//	09123456789, 9123456789, +989123456789, 00989123456789, 989123456789
//
// Non-digit separators (spaces, dashes) are tolerated.
func NormalizePhone(in string) (string, error) {
	var digitsOnly strings.Builder
	for _, r := range in {
		if r >= '0' && r <= '9' {
			digitsOnly.WriteRune(r)
		}
	}
	d := digitsOnly.String()

	switch {
	case strings.HasPrefix(d, "0098"):
		d = d[4:]
	case strings.HasPrefix(d, "98"):
		d = d[2:]
	case strings.HasPrefix(d, "0"):
		d = d[1:]
	}

	// After stripping the country/trunk prefix an Iranian mobile is exactly
	// 10 digits beginning with 9 (e.g. 9123456789).
	if len(d) != 10 || d[0] != '9' {
		return "", ErrInvalidPhone
	}
	return "+98" + d, nil
}
