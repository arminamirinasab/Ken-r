package pairing

import (
	"context"
	"testing"
	"time"

	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/ports"
)

// --- in-memory fakes ---

type memInvites struct{ m map[string]pair.Invite }

func newMemInvites() *memInvites { return &memInvites{m: map[string]pair.Invite{}} }

func (r *memInvites) Create(_ context.Context, inv pair.Invite) error {
	if _, ok := r.m[inv.Code]; ok {
		return ports.ErrCodeCollision
	}
	r.m[inv.Code] = inv
	return nil
}
func (r *memInvites) GetByCode(_ context.Context, code string) (pair.Invite, error) {
	inv, ok := r.m[code]
	if !ok {
		return pair.Invite{}, pair.ErrInviteNotFound
	}
	return inv, nil
}
func (r *memInvites) MarkAccepted(_ context.Context, code, by, pairID string) error {
	inv, ok := r.m[code]
	if !ok {
		return pair.ErrInviteNotFound
	}
	inv.Status = pair.InviteAccepted
	inv.AcceptedBy = by
	inv.PairID = pairID
	r.m[code] = inv
	return nil
}

type memPairs struct {
	m  map[string]pair.Pair
	id int
}

func newMemPairs() *memPairs { return &memPairs{m: map[string]pair.Pair{}} }

func (r *memPairs) Create(_ context.Context, p pair.Pair) (pair.Pair, error) {
	r.id++
	p.ID = "pair-" + string(rune('0'+r.id))
	r.m[p.ID] = p
	return p, nil
}
func (r *memPairs) GetByID(_ context.Context, id string) (pair.Pair, error) {
	p, ok := r.m[id]
	if !ok {
		return pair.Pair{}, pair.ErrPairNotFound
	}
	return p, nil
}
func (r *memPairs) GetActiveByUser(_ context.Context, userID string) (pair.Pair, error) {
	for _, p := range r.m {
		if p.Status == pair.PairActive && p.Contains(userID) {
			return p, nil
		}
	}
	return pair.Pair{}, pair.ErrPairNotFound
}
func (r *memPairs) Disconnect(_ context.Context, id string) error {
	p, ok := r.m[id]
	if !ok {
		return pair.ErrPairNotFound
	}
	p.Status = pair.PairDisconnected
	r.m[id] = p
	return nil
}

func newService() *Service {
	return New(nil, newMemInvites(), newMemPairs(), func() time.Time { return time.Unix(1_700_000_000, 0) })
}

// --- tests ---

func TestPairingHappyPath(t *testing.T) {
	ctx := context.Background()
	svc := newService()

	inv, err := svc.CreateInvite(ctx, "alice")
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	p, err := svc.AcceptInvite(ctx, inv.Code, "bob")
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if !p.Contains("alice") || !p.Contains("bob") {
		t.Fatalf("pair missing members: %+v", p)
	}

	space, err := svc.Space(ctx, "alice")
	if err != nil || space.ID != p.ID {
		t.Fatalf("Space: %+v err=%v", space, err)
	}
	if err := svc.Disconnect(ctx, "bob"); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	if _, err := svc.Space(ctx, "alice"); err != pair.ErrPairNotFound {
		t.Fatalf("expected no active pair after disconnect, got %v", err)
	}
}

func TestCannotAcceptOwnInvite(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	inv, _ := svc.CreateInvite(ctx, "alice")
	if _, err := svc.AcceptInvite(ctx, inv.Code, "alice"); err != pair.ErrSelfPairing {
		t.Fatalf("got %v, want ErrSelfPairing", err)
	}
}

func TestCannotInviteWhenAlreadyPaired(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	inv, _ := svc.CreateInvite(ctx, "alice")
	if _, err := svc.AcceptInvite(ctx, inv.Code, "bob"); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CreateInvite(ctx, "alice"); err != pair.ErrAlreadyPaired {
		t.Fatalf("got %v, want ErrAlreadyPaired", err)
	}
}

func TestAcceptInvalidCode(t *testing.T) {
	ctx := context.Background()
	svc := newService()
	if _, err := svc.AcceptInvite(ctx, "KENAR-ZZZZ", "bob"); err != pair.ErrInviteNotFound {
		t.Fatalf("got %v, want ErrInviteNotFound", err)
	}
}
