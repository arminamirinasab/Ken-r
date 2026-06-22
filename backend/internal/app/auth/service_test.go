package auth

import (
	"context"
	"testing"
	"time"

	"github.com/kenar/backend/internal/adapters/memory"
	authdomain "github.com/kenar/backend/internal/domain/auth"
	"github.com/kenar/backend/internal/platform/i18n"
)

// captureSMS records the last message "sent" so tests can read the OTP code.
type captureSMS struct {
	phone, body string
}

func (c *captureSMS) Send(_ context.Context, phone, message string) error {
	c.phone, c.body = phone, message
	return nil
}

func newAuth(t *testing.T, now func() time.Time) (*Service, *captureSMS) {
	t.Helper()
	bundle, err := i18n.Load("../../../i18n", "fa")
	if err != nil {
		t.Fatalf("load i18n: %v", err)
	}
	sms := &captureSMS{}
	svc := New(memory.NewUsers(), memory.NewOTPs(), memory.NewSessions(now), sms, bundle, now)
	return svc, sms
}

func TestRequestAndVerifyOTP(t *testing.T) {
	ctx := context.Background()
	now := func() time.Time { return time.Unix(1_700_000_000, 0) }
	svc, sms := newAuth(t, now)

	phone, err := svc.RequestOTP(ctx, "09123456789", "fa")
	if err != nil {
		t.Fatalf("RequestOTP: %v", err)
	}
	if phone != "+989123456789" {
		t.Fatalf("normalized phone = %q", phone)
	}
	// The SMS body carries the code; recover it for the verify step.
	code := sms.body[len(sms.body)-authdomain.OTPDigits:]

	token, user, err := svc.VerifyOTP(ctx, "0912 345 6789", code, "Ava")
	if err != nil {
		t.Fatalf("VerifyOTP: %v", err)
	}
	if token == "" || user.ID == "" {
		t.Fatalf("expected token and user, got token=%q user=%+v", token, user)
	}
	if user.DisplayName != "Ava" {
		t.Fatalf("display name = %q", user.DisplayName)
	}

	// The session authenticates.
	uid, err := svc.Authenticate(ctx, token)
	if err != nil || uid != user.ID {
		t.Fatalf("Authenticate: uid=%q err=%v", uid, err)
	}
}

func TestVerifyOTPWrongCode(t *testing.T) {
	ctx := context.Background()
	now := func() time.Time { return time.Unix(1_700_000_000, 0) }
	svc, _ := newAuth(t, now)

	if _, err := svc.RequestOTP(ctx, "09123456789", "fa"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.VerifyOTP(ctx, "09123456789", "000000", ""); err != authdomain.ErrOTPInvalid {
		t.Fatalf("got %v, want ErrOTPInvalid", err)
	}
}

func TestVerifyOTPExpired(t *testing.T) {
	ctx := context.Background()
	current := time.Unix(1_700_000_000, 0)
	now := func() time.Time { return current }
	svc, sms := newAuth(t, now)

	if _, err := svc.RequestOTP(ctx, "09123456789", "fa"); err != nil {
		t.Fatal(err)
	}
	code := sms.body[len(sms.body)-authdomain.OTPDigits:]

	current = current.Add(OTPTTL + time.Second) // advance past expiry
	if _, _, err := svc.VerifyOTP(ctx, "09123456789", code, ""); err != authdomain.ErrOTPExpired {
		t.Fatalf("got %v, want ErrOTPExpired", err)
	}
}

func TestVerifyReturnsSameUserOnSecondLogin(t *testing.T) {
	ctx := context.Background()
	now := func() time.Time { return time.Unix(1_700_000_000, 0) }
	svc, sms := newAuth(t, now)

	login := func() string {
		if _, err := svc.RequestOTP(ctx, "09123456789", "fa"); err != nil {
			t.Fatal(err)
		}
		code := sms.body[len(sms.body)-authdomain.OTPDigits:]
		_, u, err := svc.VerifyOTP(ctx, "09123456789", code, "")
		if err != nil {
			t.Fatal(err)
		}
		return u.ID
	}

	if first, second := login(), login(); first != second {
		t.Fatalf("user id changed across logins: %q vs %q", first, second)
	}
}
