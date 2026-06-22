// Package widgets implements the shared-widget use cases: an author posts a new
// payload for a widget kind, which is persisted as the pair's latest state and
// published to the real-time fan-out so the partner's widget updates. Reads
// return the current state of both partners. Depends only on ports + domain.
package widgets

import (
	"context"
	"fmt"
	"time"

	"github.com/kenar/backend/internal/domain/widget"
	"github.com/kenar/backend/internal/ports"
)

// Clock lets tests control time; production uses time.Now.
type Clock func() time.Time

// Service holds the widget use cases.
type Service struct {
	repo  ports.WidgetRepo
	pairs ports.PairRepo
	pub   ports.EventPublisher
	now   Clock
}

// New constructs a widgets Service.
func New(repo ports.WidgetRepo, pairs ports.PairRepo, pub ports.EventPublisher, clock Clock) *Service {
	if clock == nil {
		clock = time.Now
	}
	return &Service{repo: repo, pairs: pairs, pub: pub, now: clock}
}

// Set validates and stores authorID's latest payload for a widget kind within
// their active pair, then publishes a change event for real-time delivery.
func (s *Service) Set(ctx context.Context, authorID string, kind widget.Kind, payload []byte, meta map[string]string) (widget.State, error) {
	if !kind.Valid() {
		return widget.State{}, widget.ErrUnknownKind
	}
	if len(payload) == 0 {
		return widget.State{}, widget.ErrEmptyPayload
	}
	if len(payload) > widget.MaxPayloadBytes {
		return widget.State{}, widget.ErrPayloadTooLarge
	}

	p, err := s.pairs.GetActiveByUser(ctx, authorID)
	if err != nil {
		return widget.State{}, err // pair.ErrPairNotFound when unpaired
	}

	saved, err := s.repo.Save(ctx, widget.State{
		PairID:      p.ID,
		Kind:        kind,
		AuthorID:    authorID,
		Payload:     payload,
		PayloadMeta: meta,
		UpdatedAt:   s.now(),
	})
	if err != nil {
		return widget.State{}, fmt.Errorf("widgets: save state: %w", err)
	}

	ev := ports.Event{
		PairID:      p.ID,
		WidgetKind:  string(kind),
		AuthorID:    authorID,
		Version:     saved.Version,
		PayloadMeta: meta["url"], // non-sensitive hint, e.g. media URL
	}
	if err := s.pub.Publish(ctx, ev); err != nil {
		return widget.State{}, fmt.Errorf("widgets: publish event: %w", err)
	}
	return saved, nil
}

// Latest returns the current state of every author for a widget kind in the
// caller's active pair (e.g. both partners' moods).
func (s *Service) Latest(ctx context.Context, userID string, kind widget.Kind) ([]widget.State, error) {
	if !kind.Valid() {
		return nil, widget.ErrUnknownKind
	}
	p, err := s.pairs.GetActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.repo.LatestByPairKind(ctx, p.ID, kind)
}
