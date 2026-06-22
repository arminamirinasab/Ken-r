-- Kenâr initial schema.
-- Applied automatically by the postgres container on first init.
-- Source of truth for migrations lives in backend/migrations/ once the
-- migration tooling is wired; this file bootstraps a fresh local stack.

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- gen_random_uuid()

-- ---------------------------------------------------------------------------
-- Users: identified by phone (OTP auth). Locale stored for server-generated
-- user-facing text (bilingual fa/en requirement).
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone         TEXT NOT NULL UNIQUE,
    display_name  TEXT,
    locale        TEXT NOT NULL DEFAULT 'fa' CHECK (locale IN ('fa', 'en')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- Pairs: the private shared space between exactly two users.
-- Premium is PAIR-LEVEL and SERVER-AUTHORITATIVE.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS pairs (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status             TEXT NOT NULL DEFAULT 'active'
                         CHECK (status IN ('active', 'disconnected')),
    premium_until      TIMESTAMPTZ,          -- NULL = free; server checks this
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT pair_distinct_users CHECK (user_a <> user_b)
);

-- A user can belong to at most one ACTIVE pair (Phase 1 = exactly one partner).
CREATE UNIQUE INDEX IF NOT EXISTS uq_pairs_user_a_active
    ON pairs(user_a) WHERE status = 'active';
CREATE UNIQUE INDEX IF NOT EXISTS uq_pairs_user_b_active
    ON pairs(user_b) WHERE status = 'active';

-- ---------------------------------------------------------------------------
-- Invites: short human-friendly code one user shares to pair with another.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS invites (
    code         TEXT PRIMARY KEY,                 -- e.g. "KENAR-7QF3"
    inviter_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'accepted', 'revoked', 'expired')),
    accepted_by  UUID REFERENCES users(id) ON DELETE SET NULL,
    pair_id      UUID REFERENCES pairs(id) ON DELETE SET NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_invites_inviter ON invites(inviter_id);

-- ---------------------------------------------------------------------------
-- Widget state: latest payload per (pair, widget_kind, author). The server is
-- a BLIND RELAY — payload is an opaque, client-side-encrypted blob (E2E).
-- This single generic table backs every widget kind (mood, love_tap, drawing,
-- photo, countdown, ...) so new widgets need no schema change.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS widget_state (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pair_id       UUID NOT NULL REFERENCES pairs(id) ON DELETE CASCADE,
    widget_kind   TEXT NOT NULL,        -- 'mood' | 'love_tap' | 'drawing' | ...
    author_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    payload       BYTEA NOT NULL,       -- opaque E2E-encrypted blob
    payload_meta  JSONB NOT NULL DEFAULT '{}'::jsonb, -- non-sensitive hints (e.g. media url)
    version       BIGINT NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_widget_state_pair_kind
    ON widget_state(pair_id, widget_kind, created_at DESC);

-- One latest row per (pair, kind, author). The Postgres WidgetRepo upserts:
--   INSERT ... ON CONFLICT (pair_id, widget_kind, author_id)
--   DO UPDATE SET payload = EXCLUDED.payload, version = widget_state.version + 1,
--                 updated_at = now();
-- so each partner keeps exactly one current state per widget (e.g. their mood),
-- matching the in-memory WidgetRepo used in tests and local dev.
CREATE UNIQUE INDEX IF NOT EXISTS uq_widget_state_latest
    ON widget_state(pair_id, widget_kind, author_id);

-- ---------------------------------------------------------------------------
-- Device registrations: where to wake a user (Pushe token) + last seen.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS devices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    push_provider TEXT NOT NULL DEFAULT 'pushe',
    push_token    TEXT NOT NULL,
    platform      TEXT NOT NULL DEFAULT 'android',
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, push_token)
);
CREATE INDEX IF NOT EXISTS idx_devices_user ON devices(user_id);
