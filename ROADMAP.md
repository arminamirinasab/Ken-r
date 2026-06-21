# Kenâr — ROADMAP (Single Source of Truth)

> **Read this file at the start of EVERY session before doing anything.**
> Update it at the end of every working session. Never re-architect from memory.

کِنار — *"Always beside you" (همیشه کنارت)*

---

## 1. Vision

Kenâr (Persian: کِنار — "beside / next to") is a **widget-first, presence-first**
private shared space for **exactly two paired people**. It is **not a messenger**
and **not a social network**. The shared space is rendered as Android home-screen
widgets on **both** phones; a change by one person appears near-instantly on the
other's widget **without either person opening an app**.

Target feeling: *"my partner is present in my life right now, even kilometers away."*

- **Phase 1 audience:** couples, engaged, long-distance.
- **Phase 2 audience:** close friends, siblings, family.
- **Phase 3 audience:** small groups, parent–child, two-person teams.
- **Differentiation:** competitors are chat-centric / app-centric. Kenâr is
  widget-first and presence-first.

---

## 2. Hard Constraints (Iran context — non-negotiable)

- ❌ NO Firebase, NO Supabase, NO Google Play Services, NO foreign cloud SaaS.
- ✅ Push wake channel: **Pushe** (Iranian, FCM-independent), device-to-device,
  abstracted behind a `PushProvider` interface so it can be swapped.
- ✅ Everything **self-hostable on a single Iranian server**.
- ✅ Distribution: **Cafe Bazaar + Myket** (NOT Google Play).

---

## 3. Locked Tech Stack

**Android**
- Kotlin, Jetpack Compose (app UI), **Jetpack Glance** (home-screen widgets),
  Coil (images), MVVM/MVI + Clean Architecture (domain / data / presentation).

**Backend**
- **Go**, hexagonal/clean layering.
- **PostgreSQL** (primary data) + **Redis** (pub/sub, presence, ephemeral state).
- **WebSocket** for real-time when app connected; **Pushe** wakes device when killed.
- Object storage: self-hosted **MinIO** (S3-compatible) for photos/voice.
- Auth: **phone OTP** via Iranian SMS provider behind an `SMSProvider` interface
  (**Kavenegar** default).

**Security**
- End-to-end encryption of payloads (encrypt client-side; server only relays).
- Premium status is **SERVER-AUTHORITATIVE**; app never decides premium locally;
  premium content not delivered unless server confirms active subscription.
- R8 obfuscation enabled.

**Ops**
- Docker used for reproducible local + single-server deploy (see `infra/`).

---

## 4. Localization (REQUIRED — full bilingual, Definition-of-Done)

- Persian (fa) **default/primary** + English (en), both **fully** implemented.
- No hardcoded user-facing strings anywhere. Android string resources for fa+en;
  parallel i18n catalog (`backend/i18n/*.json`) for server-generated text.
- Full **RTL** for Persian, full **LTR** for English (layouts, flipped chevrons/icons,
  date/number formatting, widget layouts).
- Locale-aware dates/times/numbers; **Jalali** calendar option for Persian.
- App respects system locale + offers in-app language switch.
- **A feature is not "done" until both fa & en strings, layouts, and formatting are done.**

---

## 5. Architecture & Quality Bar

- Clean layered architecture: domain / data / presentation (Android);
  hexagonal ports & adapters (Go).
- Performance/battery: widgets passive & stateless; small payloads;
  debounce/coalesce updates; aggressively downscale images before widget render;
  respect Android widget bitmap/IPC limits.
- Handle OEM background-kill (MIUI/Xiaomi, Samsung): in-app battery-whitelist flow.
- **Real-time pipeline (the backbone):**
  `A acts → backend writes (Postgres) + publishes (Redis) → if B socket connected,
  push over WebSocket; else wake B via Pushe → B fetches state → updates Glance widget.`
  Target latency: a few seconds.
- Tests for core logic. Small functions, clear names, explicit error handling.
  No dead code. No secrets in repo (env/config).

---

## 6. Feature Plan (phased — finish MVP solidly first)

### MVP — Phase 1
1. **Pair System** — invite-code pairing; private shared space; manage/disconnect.
   Premium attaches to the **pair** (one buys, both unlock).
2. **Shared Drawing Widget** — draw on small canvas; appears on partner's widget.
3. **Mood Widget** — emotional status (happy/sad/tired/loving/angry) on partner widget.
4. **Love Tap Widget** — quick buttons (I love you / I miss you / good night /
   good morning); tap shows message on partner's widget.
5. **Shared Photo Widget** — send a moment photo; latest image shows on both.
6. **Countdown Widget** — countdown to next date / trip / anniversary / birthday.

### Suggested core features — first-class (do NOT drop)
7. **"Their World"** — partner's local time + weather + day/night state.
8. **Tap-back / reciprocity loop** — tap a received Love Tap to send "caught it" back.
9. **Presence Pulse** — both viewing at once → subtle synced pulse.
10. **Hold Hands** — both press simultaneously → shared synchronized haptic.
11. **Sealed "Open When…" messages** — pre-written, unlock on tap.
12. **Dual daily photo reveal** — daily prompt; revealed only once both posted.

### Phase 2
Memory Widget, Relationship Plant, Shared Journal, Couple Goals, Voice Moments.

### Phase 3
Lock Screen widgets, AI Companion, Relationship Timeline.

---

## 7. Monetization

Freemium. Free: one pair + base widgets. Premium: exclusive widgets, themes,
unlimited history, more storage, AI. **Premium is pair-level** (gift to the
relationship). Billing via Cafe Bazaar / Myket in-app billing; **validate
purchases server-side**; never trust a local premium flag.

---

## 8. Design Direction

Distinctive, minimal, romantic — modern & youthful, never cheesy. Warm, intimate
micro-interactions. Cohesive design tokens (color, type, spacing, motion).
Beautiful empty states. Widgets are the face of the product — gorgeous & glanceable.
Light/dark. Persian-first full RTL + polished English/LTR.

---

## 9. Repository Layout

```
/ROADMAP.md          ← this file (source of truth)
/README.md           ← setup/run instructions
/docs/               ← architecture & decision records
/infra/              ← docker-compose, postgres init, env templates
/backend/            ← Go service (hexagonal: domain/app/adapters)
  /i18n/             ← server-side fa/en message catalogs
/android/            ← Kotlin app (Compose UI + Glance widgets)
  /app/src/main/res/values/      ← Persian (default/primary)
  /app/src/main/res/values-en/   ← English
```

---

## 10. Engineering Process

- Git from first commit. Conventional Commits (feat/fix/refactor/chore/docs).
- Small, atomic, well-described commits per feature or sensible checkpoint.
- Maintain README with setup/run instructions.
- This ROADMAP is the persistent source of truth.

---

## 11. Progress Checklist

Legend: `[ ]` todo · `[~]` in progress · `[x]` done

### Foundation
- [x] Create ROADMAP.md
- [x] Git init (branch `main`)
- [~] Repo scaffold (backend + android + infra) with bilingual i18n in place
- [ ] README with setup/run instructions
- [ ] docs/ARCHITECTURE.md + ADRs
- [ ] infra: docker-compose (Postgres, Redis, MinIO), env templates
- [ ] Backend: config loader, structured logging, health endpoint
- [ ] Backend: i18n catalog loader (fa/en) + locale middleware
- [ ] Backend ports: PushProvider (Pushe), SMSProvider (Kavenegar)
- [ ] Android: gradle scaffold, Compose + Glance deps, Hilt/DI
- [ ] Android: string resources fa (default) + en; RTL config; locale switch

### Real-time backbone
- [ ] Postgres schema + migrations (users, pairs, invites, widget_state, events)
- [ ] WebSocket hub (connect/auth/presence) + Redis pub/sub fan-out
- [ ] Pushe wake path when socket absent
- [ ] Generic widget-state event model (write → publish → deliver)

### MVP Phase 1 features
- [ ] Auth: phone OTP (Kavenegar) + session/JWT
- [ ] Pair System: invite code, accept, shared space, disconnect
- [ ] Mood Widget (simplest end-to-end vertical slice first)
- [ ] Love Tap Widget (+ tap-back loop)
- [ ] Shared Drawing Widget
- [ ] Shared Photo Widget (MinIO upload + downscale)
- [ ] Countdown Widget
- [ ] "Their World" widget
- [ ] Presence Pulse / Hold Hands
- [ ] Sealed "Open When…" messages
- [ ] Dual daily photo reveal
- [ ] Premium gating (server-authoritative) + Bazaar/Myket billing validation
- [ ] OEM battery-whitelist guidance flow

---

## 12. Assumptions & Decisions Log

- **Android default locale folder = Persian.** `res/values/` holds Persian
  (primary/default fallback); `res/values-en/` holds English. Documented because
  it inverts the common English-in-`values/` convention, per the Persian-first rule.
- **App/package id:** `ir.kenar` (Iranian `.ir` namespace, matches distribution).
- **Go module path:** `github.com/kenar/backend` (placeholder; swap to real
  self-hosted VCS path when chosen).
- **Toolchain note:** Go is not installed in the current dev environment; backend
  code is scaffolded and compiled/CI-verified once the Go toolchain is provisioned.
- **E2E encryption:** server is a blind relay for payload blobs; key exchange
  between the two paired devices is a dedicated design task (see docs/ — TBD).

---

## 13. Session Log

- **2026-06-22** — Session 1: Created ROADMAP.md, git init on `main`. Began repo
  scaffold (backend/android/infra) with bilingual i18n structure. (in progress)
