package memory

import (
	"context"
	"sync"

	"github.com/kenar/backend/internal/ports"
)

// Bus is an in-process pub/sub implementing both ports.EventPublisher and
// ports.EventSubscriber. It fans events out to every subscriber of a pair.
// It stands in for the Redis pub/sub adapter (same ports) in tests and local
// single-process runs.
//
// Publish never blocks on a slow subscriber: each subscriber has a small
// buffered channel and an event is dropped for that subscriber if its buffer is
// full (the client re-fetches authoritative state on reconnect, so a missed
// nudge is not fatal — it mirrors the real "wake then fetch" model).
type Bus struct {
	mu     sync.Mutex
	subs   map[string]map[int]chan ports.Event // pairID -> subID -> chan
	nextID int
}

// NewBus constructs an empty event bus.
func NewBus() *Bus {
	return &Bus{subs: map[string]map[int]chan ports.Event{}}
}

// subBuffer is the per-subscriber backlog before events are dropped.
const subBuffer = 16

// Publish delivers ev to every current subscriber of ev.PairID.
func (b *Bus) Publish(_ context.Context, ev ports.Event) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs[ev.PairID] {
		select {
		case ch <- ev:
		default: // subscriber is behind; drop — it will re-fetch on reconnect
		}
	}
	return nil
}

// Subscribe returns a channel of events for pairID. The channel is closed and
// the subscription removed when ctx is cancelled.
func (b *Bus) Subscribe(ctx context.Context, pairID string) (<-chan ports.Event, error) {
	ch := make(chan ports.Event, subBuffer)

	b.mu.Lock()
	if b.subs[pairID] == nil {
		b.subs[pairID] = map[int]chan ports.Event{}
	}
	sid := b.nextID
	b.nextID++
	b.subs[pairID][sid] = ch
	b.mu.Unlock()

	go func() {
		<-ctx.Done()
		b.mu.Lock()
		if m := b.subs[pairID]; m != nil {
			delete(m, sid)
			if len(m) == 0 {
				delete(b.subs, pairID)
			}
		}
		close(ch)
		b.mu.Unlock()
	}()

	return ch, nil
}
