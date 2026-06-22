package memory

import (
	"context"
	"testing"
	"time"

	"github.com/kenar/backend/internal/domain/widget"
	"github.com/kenar/backend/internal/ports"
)

func TestWidgetsSaveVersioning(t *testing.T) {
	ctx := context.Background()
	repo := NewWidgets()

	first, err := repo.Save(ctx, widget.State{PairID: "p1", Kind: widget.KindMood, AuthorID: "a", Payload: []byte("x")})
	if err != nil {
		t.Fatal(err)
	}
	if first.Version != 1 || first.ID == "" {
		t.Fatalf("first save = %+v, want version 1 with id", first)
	}

	second, err := repo.Save(ctx, widget.State{PairID: "p1", Kind: widget.KindMood, AuthorID: "a", Payload: []byte("y")})
	if err != nil {
		t.Fatal(err)
	}
	if second.Version != 2 {
		t.Fatalf("second version = %d, want 2", second.Version)
	}
	if second.ID != first.ID {
		t.Fatalf("id changed on upsert: %s -> %s", first.ID, second.ID)
	}

	// A different author keeps a separate latest state.
	if _, err := repo.Save(ctx, widget.State{PairID: "p1", Kind: widget.KindMood, AuthorID: "b", Payload: []byte("z")}); err != nil {
		t.Fatal(err)
	}
	states, err := repo.LatestByPairKind(ctx, "p1", widget.KindMood)
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 2 {
		t.Fatalf("latest count = %d, want 2 (one per author)", len(states))
	}
}

func TestBusFanOut(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub1, _ := bus.Subscribe(ctx, "pair-1")
	sub2, _ := bus.Subscribe(ctx, "pair-1")
	other, _ := bus.Subscribe(ctx, "pair-2")

	ev := ports.Event{PairID: "pair-1", WidgetKind: "mood", AuthorID: "a", Version: 1}
	if err := bus.Publish(ctx, ev); err != nil {
		t.Fatal(err)
	}

	for i, ch := range []<-chan ports.Event{sub1, sub2} {
		select {
		case got := <-ch:
			if got != ev {
				t.Fatalf("sub%d got %+v, want %+v", i, got, ev)
			}
		case <-time.After(time.Second):
			t.Fatalf("sub%d did not receive event", i)
		}
	}

	select {
	case got := <-other:
		t.Fatalf("subscriber for other pair received %+v", got)
	case <-time.After(50 * time.Millisecond):
		// expected: no delivery to a different pair
	}
}

func TestBusUnsubscribeOnContextCancel(t *testing.T) {
	bus := NewBus()
	ctx, cancel := context.WithCancel(context.Background())
	ch, _ := bus.Subscribe(ctx, "pair-1")

	cancel()
	select {
	case _, open := <-ch:
		if open {
			t.Fatal("expected channel closed after context cancel")
		}
	case <-time.After(time.Second):
		t.Fatal("channel was not closed after context cancel")
	}
}
