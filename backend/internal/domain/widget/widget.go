// Package widget holds the core domain for shared widget state — the small,
// glanceable payloads one partner sends that surface on the other's home-screen
// widget. The server is a BLIND RELAY: Payload is an opaque, client-side
// (E2E) encrypted blob and is never inspected here. This package has NO
// infrastructure dependencies.
package widget

import (
	"errors"
	"time"
)

// Kind identifies which widget a piece of state belongs to. One generic state
// model backs every widget so new widgets need no schema or domain change.
type Kind string

const (
	KindMood      Kind = "mood"
	KindLoveTap   Kind = "love_tap"
	KindDrawing   Kind = "drawing"
	KindPhoto     Kind = "photo"
	KindCountdown  Kind = "countdown"
	KindTheirWorld Kind = "their_world"
)

var validKinds = map[Kind]bool{
	KindMood:       true,
	KindLoveTap:    true,
	KindDrawing:    true,
	KindPhoto:      true,
	KindCountdown:  true,
	KindTheirWorld: true,
}

// Valid reports whether k is a known widget kind.
func (k Kind) Valid() bool { return validKinds[k] }

// MaxPayloadBytes caps the encrypted blob. Widgets are tiny and must respect
// Android widget bitmap/IPC limits; large media goes to object storage and only
// its URL travels in PayloadMeta.
const MaxPayloadBytes = 64 * 1024

// Domain errors. Adapters map these to transport-level responses + i18n keys.
var (
	ErrUnknownKind     = errors.New("unknown widget kind")
	ErrEmptyPayload    = errors.New("widget payload is empty")
	ErrPayloadTooLarge = errors.New("widget payload too large")
)

// State is the latest payload a single author posted for one widget kind within
// a pair. Mood, Love Tap, etc. each keep one State per (pair, kind, author).
type State struct {
	ID          string
	PairID      string
	Kind        Kind
	AuthorID    string
	Payload     []byte            // opaque, client-encrypted (E2E) blob
	PayloadMeta map[string]string // non-sensitive hints (e.g. media URL)
	Version     int64             // monotonically increasing per (pair, kind, author)
	UpdatedAt   time.Time
}
