// Package memory provides in-memory implementations of every persistence port
// plus an in-process event bus. It backs unit/integration tests and lets the
// full server run locally with zero external infrastructure. The production
// Postgres/Redis adapters implement the same ports and swap in via wiring.
//
// Every type is safe for concurrent use.
package memory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	authdomain "github.com/kenar/backend/internal/domain/auth"
	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/domain/widget"
	"github.com/kenar/backend/internal/ports"
)

// id returns a short random hex identifier with the given prefix.
func id(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return prefix + "_" + hex.EncodeToString(b)
}

// --- Users ---

// Users is an in-memory ports.UserRepo.
type Users struct {
	mu      sync.RWMutex
	byID    map[string]pair.User
	byPhone map[string]string // phone -> id
}

// NewUsers constructs an empty user store.
func NewUsers() *Users {
	return &Users{byID: map[string]pair.User{}, byPhone: map[string]string{}}
}

func (s *Users) GetByID(_ context.Context, uid string) (pair.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[uid]
	if !ok {
		return pair.User{}, pair.ErrUserNotFound
	}
	return u, nil
}

func (s *Users) GetByPhone(_ context.Context, phone string) (pair.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	uid, ok := s.byPhone[phone]
	if !ok {
		return pair.User{}, pair.ErrUserNotFound
	}
	return s.byID[uid], nil
}

func (s *Users) Create(_ context.Context, u pair.User) (pair.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if u.ID == "" {
		u.ID = id("u")
	}
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}
	s.byID[u.ID] = u
	s.byPhone[u.Phone] = u.ID
	return u, nil
}

// --- Invites ---

// Invites is an in-memory ports.InviteRepo.
type Invites struct {
	mu sync.RWMutex
	m  map[string]pair.Invite
}

// NewInvites constructs an empty invite store.
func NewInvites() *Invites { return &Invites{m: map[string]pair.Invite{}} }

func (s *Invites) Create(_ context.Context, inv pair.Invite) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.m[inv.Code]; ok {
		return ports.ErrCodeCollision
	}
	s.m[inv.Code] = inv
	return nil
}

func (s *Invites) GetByCode(_ context.Context, code string) (pair.Invite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inv, ok := s.m[code]
	if !ok {
		return pair.Invite{}, pair.ErrInviteNotFound
	}
	return inv, nil
}

func (s *Invites) MarkAccepted(_ context.Context, code, acceptedBy, pairID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.m[code]
	if !ok {
		return pair.ErrInviteNotFound
	}
	inv.Status = pair.InviteAccepted
	inv.AcceptedBy = acceptedBy
	inv.PairID = pairID
	s.m[code] = inv
	return nil
}

// --- Pairs ---

// Pairs is an in-memory ports.PairRepo.
type Pairs struct {
	mu sync.RWMutex
	m  map[string]pair.Pair
}

// NewPairs constructs an empty pair store.
func NewPairs() *Pairs { return &Pairs{m: map[string]pair.Pair{}} }

func (s *Pairs) Create(_ context.Context, p pair.Pair) (pair.Pair, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = id("pair")
	}
	s.m[p.ID] = p
	return p, nil
}

func (s *Pairs) GetByID(_ context.Context, pid string) (pair.Pair, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.m[pid]
	if !ok {
		return pair.Pair{}, pair.ErrPairNotFound
	}
	return p, nil
}

func (s *Pairs) GetActiveByUser(_ context.Context, userID string) (pair.Pair, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, p := range s.m {
		if p.Status == pair.PairActive && p.Contains(userID) {
			return p, nil
		}
	}
	return pair.Pair{}, pair.ErrPairNotFound
}

func (s *Pairs) Disconnect(_ context.Context, pairID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[pairID]
	if !ok {
		return pair.ErrPairNotFound
	}
	p.Status = pair.PairDisconnected
	s.m[pairID] = p
	return nil
}

// --- Devices ---

// Devices is an in-memory ports.DeviceRepo.
type Devices struct {
	mu sync.RWMutex
	m  map[string][]ports.Device // userID -> devices
}

// NewDevices constructs an empty device store.
func NewDevices() *Devices { return &Devices{m: map[string][]ports.Device{}} }

func (s *Devices) Upsert(_ context.Context, userID, pushToken, provider string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, d := range s.m[userID] {
		if d.PushToken == pushToken {
			return nil // already registered
		}
	}
	s.m[userID] = append(s.m[userID], ports.Device{UserID: userID, Provider: provider, PushToken: pushToken})
	return nil
}

func (s *Devices) ListByUser(_ context.Context, userID string) ([]ports.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ports.Device, len(s.m[userID]))
	copy(out, s.m[userID])
	return out, nil
}

// --- Widgets ---

type widgetKey struct {
	pairID, kind, author string
}

// Widgets is an in-memory ports.WidgetRepo keyed by (pair, kind, author).
type Widgets struct {
	mu sync.RWMutex
	m  map[widgetKey]widget.State
}

// NewWidgets constructs an empty widget-state store.
func NewWidgets() *Widgets { return &Widgets{m: map[widgetKey]widget.State{}} }

func (s *Widgets) Save(_ context.Context, st widget.State) (widget.State, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := widgetKey{st.PairID, string(st.Kind), st.AuthorID}
	if prev, ok := s.m[k]; ok {
		st.ID = prev.ID
		st.Version = prev.Version + 1
	} else {
		st.ID = id("ws")
		st.Version = 1
	}
	s.m[k] = st
	return st, nil
}

func (s *Widgets) LatestByPairKind(_ context.Context, pairID string, kind widget.Kind) ([]widget.State, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []widget.State
	for k, st := range s.m {
		if k.pairID == pairID && k.kind == string(kind) {
			out = append(out, st)
		}
	}
	return out, nil
}

// --- OTP ---

type otpEntry struct {
	code      string
	expiresAt time.Time
}

// OTPs is an in-memory ports.OTPRepo.
type OTPs struct {
	mu sync.Mutex
	m  map[string]otpEntry
}

// NewOTPs constructs an empty OTP store.
func NewOTPs() *OTPs { return &OTPs{m: map[string]otpEntry{}} }

func (s *OTPs) Put(_ context.Context, phone, code string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[phone] = otpEntry{code: code, expiresAt: expiresAt}
	return nil
}

func (s *OTPs) Get(_ context.Context, phone string) (string, time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.m[phone]
	if !ok {
		return "", time.Time{}, authdomain.ErrOTPNotFound
	}
	return e.code, e.expiresAt, nil
}

func (s *OTPs) Delete(_ context.Context, phone string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, phone)
	return nil
}

// --- Sessions ---

type sessionEntry struct {
	userID    string
	expiresAt time.Time
}

// Sessions is an in-memory ports.SessionRepo issuing random bearer tokens.
type Sessions struct {
	mu  sync.RWMutex
	m   map[string]sessionEntry
	now func() time.Time
}

// NewSessions constructs an empty session store. clock may be nil (time.Now).
func NewSessions(clock func() time.Time) *Sessions {
	if clock == nil {
		clock = time.Now
	}
	return &Sessions{m: map[string]sessionEntry{}, now: clock}
}

func (s *Sessions) Create(_ context.Context, userID string, expiresAt time.Time) (string, error) {
	tok := id("s")
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[tok] = sessionEntry{userID: userID, expiresAt: expiresAt}
	return tok, nil
}

func (s *Sessions) UserID(_ context.Context, token string) (string, error) {
	s.mu.RLock()
	e, ok := s.m[token]
	s.mu.RUnlock()
	if !ok {
		return "", authdomain.ErrSessionInvalid
	}
	if !s.now().Before(e.expiresAt) {
		s.mu.Lock()
		delete(s.m, token)
		s.mu.Unlock()
		return "", authdomain.ErrSessionInvalid
	}
	return e.userID, nil
}

func (s *Sessions) Delete(_ context.Context, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, token)
	return nil
}
