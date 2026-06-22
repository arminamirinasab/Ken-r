package widgets

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/kenar/backend/internal/adapters/memory"
	"github.com/kenar/backend/internal/domain/pair"
	"github.com/kenar/backend/internal/domain/widget"
)

// fixedClock is a stable timestamp for deterministic tests.
func fixedClock() time.Time { return time.Unix(1_700_000_000, 0) }

// newWidgets wires the service to in-memory adapters and returns the bus so a
// test can assert an event was published.
func newWidgets(t *testing.T) (*Service, *memory.Pairs, *memory.Bus) {
	t.Helper()
	pairs := memory.NewPairs()
	bus := memory.NewBus()
	svc := New(memory.NewWidgets(), pairs, bus, fixedClock)
	return svc, pairs, bus
}

func TestSetPublishesEventAndPersists(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc, pairs, bus := newWidgets(t)

	p, _ := pairs.Create(ctx, pair.Pair{UserA: "alice", UserB: "bob", Status: pair.PairActive})

	// Subscribe before acting so the published event is observed.
	events, _ := bus.Subscribe(ctx, p.ID)

	st, err := svc.Set(ctx, "alice", widget.KindMood, []byte("happy-blob"), map[string]string{"url": "u1"})
	if err != nil {
		t.Fatalf("Set: %v", err)
	}
	if st.Version != 1 || st.PairID != p.ID {
		t.Fatalf("saved state = %+v", st)
	}

	select {
	case ev := <-events:
		if ev.PairID != p.ID || ev.WidgetKind != string(widget.KindMood) || ev.AuthorID != "alice" || ev.PayloadMeta != "u1" {
			t.Fatalf("unexpected event %+v", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("no event published")
	}

	got, err := svc.Latest(ctx, "bob", widget.KindMood)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || !bytes.Equal(got[0].Payload, []byte("happy-blob")) {
		t.Fatalf("Latest = %+v", got)
	}
}

func TestSetValidations(t *testing.T) {
	ctx := context.Background()
	svc, pairs, _ := newWidgets(t)
	_, _ = pairs.Create(ctx, pair.Pair{UserA: "alice", UserB: "bob", Status: pair.PairActive})

	if _, err := svc.Set(ctx, "alice", widget.Kind("bogus"), []byte("x"), nil); err != widget.ErrUnknownKind {
		t.Fatalf("unknown kind: got %v", err)
	}
	if _, err := svc.Set(ctx, "alice", widget.KindMood, nil, nil); err != widget.ErrEmptyPayload {
		t.Fatalf("empty payload: got %v", err)
	}
	big := make([]byte, widget.MaxPayloadBytes+1)
	if _, err := svc.Set(ctx, "alice", widget.KindMood, big, nil); err != widget.ErrPayloadTooLarge {
		t.Fatalf("too large: got %v", err)
	}
}

func TestSetRequiresActivePair(t *testing.T) {
	ctx := context.Background()
	svc, _, _ := newWidgets(t)
	if _, err := svc.Set(ctx, "lonely", widget.KindMood, []byte("x"), nil); err != pair.ErrPairNotFound {
		t.Fatalf("got %v, want ErrPairNotFound", err)
	}
}
