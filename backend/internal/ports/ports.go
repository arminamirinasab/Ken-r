// Package ports declares the interfaces (hexagonal "ports") the application
// core depends on. Adapters (postgres, redis, pushe, kavenegar) implement them.
// The core never imports an adapter — only this package.
package ports

import (
	"context"
	"errors"
	"time"

	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/domain/widget"
)

// ErrCodeCollision is returned by InviteRepo.Create when the generated code
// already exists, signalling the pairing service to retry with a fresh code.
var ErrCodeCollision = errors.New("invite code collision")

// --- Persistence ports ---

// UserRepo persists users (phone-identified accounts).
type UserRepo interface {
	GetByID(ctx context.Context, id string) (pair.User, error)
	GetByPhone(ctx context.Context, phone string) (pair.User, error)
	Create(ctx context.Context, u pair.User) (pair.User, error)
}

// InviteRepo persists pairing invite codes.
type InviteRepo interface {
	Create(ctx context.Context, inv pair.Invite) error
	GetByCode(ctx context.Context, code string) (pair.Invite, error)
	MarkAccepted(ctx context.Context, code, acceptedBy, pairID string) error
}

// PairRepo persists pairs (the two-person shared spaces).
type PairRepo interface {
	Create(ctx context.Context, p pair.Pair) (pair.Pair, error)
	GetByID(ctx context.Context, id string) (pair.Pair, error)
	GetActiveByUser(ctx context.Context, userID string) (pair.Pair, error)
	Disconnect(ctx context.Context, pairID string) error
}

// DeviceRepo persists per-user device push tokens used to wake the partner.
type DeviceRepo interface {
	Upsert(ctx context.Context, userID, pushToken, provider string) error
	ListByUser(ctx context.Context, userID string) ([]Device, error)
}

// Device is a registered push target for a user.
type Device struct {
	UserID    string
	Provider  string
	PushToken string
}

// WidgetRepo persists the latest widget state per (pair, kind, author). The
// server stores the opaque E2E blob without inspecting it.
type WidgetRepo interface {
	// Save upserts the author's state for a widget kind, assigning an ID on
	// first write and bumping Version on each subsequent write.
	Save(ctx context.Context, st widget.State) (widget.State, error)
	// LatestByPairKind returns the current state of every author in the pair
	// for the given widget kind (e.g. both partners' moods).
	LatestByPairKind(ctx context.Context, pairID string, kind widget.Kind) ([]widget.State, error)
}

// OTPRepo stores at most one pending login code per phone, with an expiry.
type OTPRepo interface {
	Put(ctx context.Context, phone, code string, expiresAt time.Time) error
	// Get returns auth.ErrOTPNotFound when no code is pending for the phone.
	Get(ctx context.Context, phone string) (code string, expiresAt time.Time, err error)
	Delete(ctx context.Context, phone string) error
}

// SessionRepo issues and resolves opaque bearer session tokens.
type SessionRepo interface {
	Create(ctx context.Context, userID string, expiresAt time.Time) (token string, err error)
	// UserID returns auth.ErrSessionInvalid when the token is unknown/expired.
	UserID(ctx context.Context, token string) (userID string, err error)
	Delete(ctx context.Context, token string) error
}

// --- Provider ports (swappable per Iran-context constraints) ---

// SMSProvider sends OTP/SMS. Default adapter: Kavenegar.
type SMSProvider interface {
	Send(ctx context.Context, phone, message string) error
}

// PushProvider wakes a device when its socket is not connected.
// Default adapter: Pushe (Iranian, FCM-independent).
type PushProvider interface {
	// Wake sends a silent, data-only notification to nudge the device to
	// fetch fresh widget state. Payload is small and non-sensitive.
	Wake(ctx context.Context, pushToken string, data map[string]string) error
}

// --- Real-time backbone ports ---

// Event is the unit published when a pair's shared state changes. Subscribers
// (the WebSocket hub) fan it out to connected members; absent members are woken
// via PushProvider. The payload is an opaque, client-encrypted blob (E2E).
type Event struct {
	PairID      string `json:"pair_id"`
	WidgetKind  string `json:"widget_kind"`
	AuthorID    string `json:"author_id"`
	Version     int64  `json:"version"`
	PayloadMeta string `json:"payload_meta,omitempty"` // non-sensitive hint, e.g. media URL
}

// EventPublisher publishes pair events to the real-time fan-out (Redis pub/sub).
type EventPublisher interface {
	Publish(ctx context.Context, ev Event) error
}

// EventSubscriber receives pair events for delivery to connected clients.
type EventSubscriber interface {
	// Subscribe returns a channel of events for the given pair. The channel is
	// closed when ctx is cancelled.
	Subscribe(ctx context.Context, pairID string) (<-chan Event, error)
}
