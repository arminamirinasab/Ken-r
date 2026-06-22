// Package pairing implements the Pair System use cases: creating an invite,
// accepting one to form a pair, inspecting the shared space, and disconnecting.
// It depends only on ports — never on concrete adapters.
package pairing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/ports"
)

// InviteTTL is how long a freshly created invite code stays usable.
const InviteTTL = 24 * time.Hour

// codeRetries bounds collision retries when generating a unique invite code.
const codeRetries = 5

// Clock lets tests control time; production uses time.Now.
type Clock func() time.Time

// Service holds the pairing use cases.
type Service struct {
	users   ports.UserRepo
	invites ports.InviteRepo
	pairs   ports.PairRepo
	now     Clock
}

// New constructs a pairing Service.
func New(users ports.UserRepo, invites ports.InviteRepo, pairs ports.PairRepo, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{users: users, invites: invites, pairs: pairs, now: clock}
}

// CreateInvite generates a unique, time-limited invite code for inviterID.
// A user already in an active pair cannot create an invite.
func (s *Service) CreateInvite(ctx context.Context, inviterID string) (pair.Invite, error) {
	if _, err := s.pairs.GetActiveByUser(ctx, inviterID); err == nil {
		return pair.Invite{}, pair.ErrAlreadyPaired
	} else if !errors.Is(err, pair.ErrPairNotFound) {
		return pair.Invite{}, fmt.Errorf("pairing: check active pair: %w", err)
	}

	now := s.now()
	for attempt := 0; attempt < codeRetries; attempt++ {
		code, err := pair.NewInviteCode()
		if err != nil {
			return pair.Invite{}, fmt.Errorf("pairing: gen code: %w", err)
		}
		inv := pair.Invite{
			Code:      code,
			InviterID: inviterID,
			Status:    pair.InvitePending,
			ExpiresAt: now.Add(InviteTTL),
			CreatedAt: now,
		}
		err = s.invites.Create(ctx, inv)
		if err == nil {
			return inv, nil
		}
		if errors.Is(err, ports.ErrCodeCollision) {
			continue // extraordinarily rare; retry with a new code
		}
		return pair.Invite{}, fmt.Errorf("pairing: persist invite: %w", err)
	}
	return pair.Invite{}, fmt.Errorf("pairing: could not allocate unique code after %d tries", codeRetries)
}

// AcceptInvite consumes a pending invite and forms an active pair between the
// inviter and accepterID. Both users must be free of active pairs.
func (s *Service) AcceptInvite(ctx context.Context, rawCode, accepterID string) (pair.Pair, error) {
	code := pair.NormalizeCode(rawCode)
	inv, err := s.invites.GetByCode(ctx, code)
	if err != nil {
		return pair.Pair{}, err // typically pair.ErrInviteNotFound
	}
	if err := inv.IsUsable(s.now()); err != nil {
		return pair.Pair{}, err
	}
	if inv.InviterID == accepterID {
		return pair.Pair{}, pair.ErrSelfPairing
	}

	if err := s.assertFree(ctx, inv.InviterID); err != nil {
		return pair.Pair{}, err
	}
	if err := s.assertFree(ctx, accepterID); err != nil {
		return pair.Pair{}, err
	}

	now := s.now()
	created, err := s.pairs.Create(ctx, pair.Pair{
		UserA:     inv.InviterID,
		UserB:     accepterID,
		Status:    pair.PairActive,
		CreatedAt: now,
	})
	if err != nil {
		return pair.Pair{}, fmt.Errorf("pairing: create pair: %w", err)
	}
	if err := s.invites.MarkAccepted(ctx, code, accepterID, created.ID); err != nil {
		return pair.Pair{}, fmt.Errorf("pairing: mark invite accepted: %w", err)
	}
	return created, nil
}

// Space returns the active pair a user belongs to.
func (s *Service) Space(ctx context.Context, userID string) (pair.Pair, error) {
	return s.pairs.GetActiveByUser(ctx, userID)
}

// Disconnect ends the active pairing for the given member.
func (s *Service) Disconnect(ctx context.Context, userID string) error {
	p, err := s.pairs.GetActiveByUser(ctx, userID)
	if err != nil {
		return err
	}
	if !p.Contains(userID) {
		return pair.ErrNotMember
	}
	return s.pairs.Disconnect(ctx, p.ID)
}

func (s *Service) assertFree(ctx context.Context, userID string) error {
	_, err := s.pairs.GetActiveByUser(ctx, userID)
	switch {
	case err == nil:
		return pair.ErrAlreadyPaired
	case errors.Is(err, pair.ErrPairNotFound):
		return nil
	default:
		return fmt.Errorf("pairing: check free %s: %w", userID, err)
	}
}
