// Package pair holds the core domain for the Pair System: the private shared
// space between exactly two people. This package has NO infrastructure
// dependencies — pure entities, value objects, and domain errors.
package pair

import (
	"errors"
	"time"
)

// Domain errors. Adapters map these to transport-level responses + i18n keys.
var (
	ErrAlreadyPaired  = errors.New("user already in an active pair")
	ErrInviteNotFound = errors.New("invite not found")
	ErrInviteExpired  = errors.New("invite expired")
	ErrInviteUsed     = errors.New("invite already used")
	ErrSelfPairing    = errors.New("cannot accept your own invite")
	ErrPairNotFound   = errors.New("pair not found")
	ErrNotMember      = errors.New("user is not a member of this pair")
	ErrUserNotFound   = errors.New("user not found")
)

// Locale is the user's preferred language. Persian is primary.
type Locale string

const (
	LocaleFa Locale = "fa"
	LocaleEn Locale = "en"
)

// User identified by phone number (OTP auth).
type User struct {
	ID          string
	Phone       string
	DisplayName string
	Locale      Locale
	CreatedAt   time.Time
}

// PairStatus is the lifecycle state of a pair.
type PairStatus string

const (
	PairActive       PairStatus = "active"
	PairDisconnected PairStatus = "disconnected"
)

// Pair is the shared space between two users. Premium attaches to the pair.
type Pair struct {
	ID           string
	UserA        string
	UserB        string
	Status       PairStatus
	PremiumUntil *time.Time
	CreatedAt    time.Time
}

// Contains reports whether userID is a member of the pair.
func (p Pair) Contains(userID string) bool {
	return p.UserA == userID || p.UserB == userID
}

// Partner returns the other member's id for a given member.
func (p Pair) Partner(userID string) (string, bool) {
	switch userID {
	case p.UserA:
		return p.UserB, true
	case p.UserB:
		return p.UserA, true
	default:
		return "", false
	}
}

// IsPremium reports whether the pair has active premium at instant t.
// Premium is SERVER-AUTHORITATIVE — this is evaluated server-side only.
func (p Pair) IsPremium(t time.Time) bool {
	return p.PremiumUntil != nil && p.PremiumUntil.After(t)
}

// InviteStatus is the lifecycle state of an invite code.
type InviteStatus string

const (
	InvitePending  InviteStatus = "pending"
	InviteAccepted InviteStatus = "accepted"
	InviteRevoked  InviteStatus = "revoked"
	InviteExpired  InviteStatus = "expired"
)

// Invite is a short, shareable code one user creates to pair with another.
type Invite struct {
	Code       string
	InviterID  string
	Status     InviteStatus
	AcceptedBy string
	PairID     string
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// IsUsable reports whether the invite can still be accepted at instant t.
func (i Invite) IsUsable(t time.Time) error {
	if i.Status != InvitePending {
		return ErrInviteUsed
	}
	if !t.Before(i.ExpiresAt) {
		return ErrInviteExpired
	}
	return nil
}
