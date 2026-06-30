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
- [x] Repo scaffold (backend + android + infra) with bilingual i18n in place
- [x] README with setup/run instructions
- [x] docs/ARCHITECTURE.md (ADRs still TODO)
- [x] infra: docker-compose (Postgres, Redis, MinIO), env templates
- [x] Backend: config loader, structured logging, health endpoint
- [x] Backend: i18n catalog loader (fa/en) + locale middleware
- [x] Backend ports: PushProvider (Pushe), SMSProvider (Kavenegar) — interfaces defined
- [x] Android: gradle scaffold, Compose + Glance deps, Hilt/DI
- [x] Android: string resources fa (default) + en; RTL config; locale switch

### Real-time backbone
- [~] Postgres schema (users, pairs, invites, widget_state, devices) — init SQL done (+ widget upsert unique index); migration tooling TODO
- [~] Real-time hub + pub/sub fan-out — in-process event Bus (ports.EventPublisher/Subscriber) + SSE delivery stream done & tested; Redis adapter + WebSocket transport TODO (blocked on module fetch)
- [ ] Pushe wake path when socket absent
- [x] Generic widget-state event model — `ports.Event` + `widget_state` table + publish→deliver wiring (Set → Bus.Publish → SSE) done & tested

### MVP Phase 1 features
- [~] Auth: phone OTP + session — service (request/verify), Iranian phone normalize, session tokens, HTTP handlers, bearer middleware done & tested; Kavenegar SMS adapter TODO (dev uses log provider)
- [~] Pair System: invite code, accept, shared space, disconnect — domain + use cases + HTTP handlers + in-memory repos done & tested; Postgres repos TODO
- [~] Mood Widget — backend end-to-end slice (set/get + event publish + SSE deliver) done & tested; Android Glance wiring TODO
- [~] Love Tap Widget (+ tap-back loop) — Android domain model + Glance
  widget shell with quick actions/tap-back queue done; sync upload/delivery
  wiring TODO with Android data layer
- [~] Shared Drawing Widget — Android compact drawing wire model + passive
  Glance renderer done; in-app drawing canvas and sync upload/delivery TODO
- [~] Shared Photo Widget (MinIO upload + downscale) — Android photo metadata
  wire model + downscaled cached Glance renderer done; MinIO upload/download
  sync TODO
- [~] Countdown Widget — Android countdown snapshot wire model + passive Glance
  renderer done; date editor and daily sync refresh TODO
- [~] "Their World" widget — Android partner world snapshot model + passive
  Glance renderer done; location/weather sync and privacy UX TODO
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
- **Toolchain note:** Go 1.26 is now provisioned; backend builds, vets, and tests
  green locally. External module fetch (pgx, redis, websocket, minio) is currently
  blocked (TLS timeout to sum.golang.org), so those adapters are deferred and the
  app runs on stdlib-only **in-memory adapters + in-process event bus** meanwhile.
- **Real-time transport:** SSE (`GET /v1/stream`, stdlib) is the interim delivery
  channel that closes the publish→deliver loop today. A WebSocket adapter sharing
  the same event Bus replaces/augments it for bidirectional presence features once
  the module fetch is unblocked. Transport is an adapter detail behind the Bus
  ports, so swapping it does not touch the use cases.
- **Local/dev wiring:** `cmd/server` wires in-memory repos + bus + a log-only SMS
  provider (logs the OTP — DEV ONLY). Production swaps in Postgres/Redis/Kavenegar
  implementing the same ports.
- **E2E encryption:** server is a blind relay for payload blobs; key exchange
  between the two paired devices is a dedicated design task (see docs/ — TBD).

---

## 13. Session Log

- **2026-06-22** — Session 1: Created ROADMAP.md, git init on `main`. Completed
  foundation scaffold:
  - infra: docker-compose (Postgres/Redis/MinIO/backend), `.env.example`,
    Postgres init schema (users, pairs, invites, widget_state, devices).
  - backend (Go, stdlib-only foundation): config, structured logging, bilingual
    i18n bundle (+tests), `domain/pair` entities & invite codes (+tests),
    `ports` (repos, SMS/Push/event interfaces), `app/pairing` use cases
    (+tests), `adapters/httpapi` router + locale middleware + health/locale
    routes, Dockerfile.
  - android: Gradle (Kotlin DSL + version catalog), Compose + Glance + Hilt,
    R8 release config, fa-default/en string resources, RTL + locales_config,
    in-app LocaleManager, theme tokens, bilingual landing, Mood Glance widget
    vertical slice (+ Mood unit test), adaptive icon.
  - docs: README + docs/ARCHITECTURE.md.
  - 4 commits on `main`.
  **Next session:** pick up the real-time backbone — wire Postgres repos for the
  pairing use cases, add HTTP handlers for invite create/accept/disconnect, then
  the Redis pub/sub + WebSocket hub and the Mood end-to-end slice (act → publish
  → deliver → Glance update). Then phone OTP auth.

- **2026-06-23** — Session 2: Go 1.26 provisioned (toolchain unblocked). Built the
  real-time backbone + auth + pairing API end-to-end, stdlib-only, fully tested.
  - android: committed the polished bilingual landing screen (gradient backdrop,
    brand mark, fa/en segmented language toggle), expanded Material color tokens,
    and the Gradle wrapper.
  - backend domain: `domain/widget` (Kind, State, payload caps), `domain/auth`
    (errors, OTP gen, Iranian phone normalize), `pair.ErrUserNotFound`.
  - backend ports: `WidgetRepo`, `OTPRepo`, `SessionRepo`; moved `ErrCodeCollision`
    into ports.
  - adapters: `adapters/memory` (in-memory Users/Invites/Pairs/Devices/Widgets/
    OTP/Sessions + in-process pub/sub `Bus`), `adapters/sms` log provider (dev).
  - use cases: `app/auth` (RequestOTP/VerifyOTP/Authenticate), `app/widgets`
    (Set→persist+publish / Latest).
  - httpapi: bearer-session middleware, domain-error→HTTP+i18n mapping, handlers
    for auth/pairing/devices/widgets, and an SSE `/v1/stream` delivery endpoint;
    full httptest flow incl. SSE event delivery.
  - i18n: added otp/session/widget/invite keys to fa + en.
  - infra: widget upsert unique index.
  - wired `cmd/server` on in-memory adapters; `go build`/`vet`/`test` all green.
  **Next session:** Postgres + Redis adapters (once module fetch is unblocked) to
  replace the in-memory repos and bus; Kavenegar SMS adapter; WebSocket transport
  alongside SSE; then wire the Android Mood Glance widget to the new endpoints
  (otp login → pair → set/get mood → stream-driven widget refresh).

- **2026-06-30** — Session 3: Added the Android Love Tap widget vertical slice
  without requiring local build tooling:
  - android: `LoveTap` domain model with stable wire values and bilingual labels;
    Glance `LoveTapWidget` registered in the manifest with widget metadata and a
    picker preview.
  - tap-back loop: widget can queue primary quick actions and `caught_it` replies
    into Glance state for the future sync layer to encrypt/upload through the
    existing generic `love_tap` widget backend kind.
  - tests: added `LoveTapTest` for wire round-trips, uniqueness, and tap-back
    exclusion from primary actions.
  **Next session:** wire the Android data/sync layer so queued Love Taps and Mood
  updates are uploaded/fetched through the existing auth/pair/widget endpoints,
  then mark Mood/Love Tap complete once real delivery refreshes Glance state.

- **2026-06-30** — Session 4: Added the Android Shared Drawing widget vertical
  slice:
  - android domain: `SharedDrawing`, `DrawingStroke`, and `DrawingPoint` with a
    compact bounded wire format suitable for the existing generic `drawing`
    widget endpoint.
  - android widget: passive Glance `DrawingWidget` registered in the manifest,
    with a small bitmap renderer for partner sketches plus widget picker
    metadata/preview.
  - localization/tests: added fa/en drawing strings and `SharedDrawingTest` for
    wire round-trips, malformed payload handling, coordinate clamping, and empty
    drawings.
  **Next session:** add the in-app drawing canvas + Android sync writer so local
  sketches are encrypted/uploaded and partner payloads refresh Glance state.

- **2026-06-30** — Session 5: Added the Android Shared Photo widget vertical
  slice:
  - android domain: `SharedPhoto` metadata model with a compact wire format for
    MinIO object keys, preview dimensions, optional captions, and content hashes.
  - android widget: passive Glance `PhotoWidget` registered in the manifest,
    rendering only a local cached/downscaled bitmap path written by the future
    sync layer.
  - downscale guardrails/tests: added `PhotoWidgetImageLoader` sample-size logic,
    fa/en strings, widget picker metadata/preview, and tests for photo metadata
    validation plus widget bitmap sampling.
  **Next session:** implement the Android photo capture/pick + sync handoff, then
  the MinIO backend adapter when external module fetch is available.

- **2026-06-30** — Session 6: Added the Android Countdown widget vertical slice:
  - android domain: `Countdown` snapshot model with bounded wire format for
    title, target epoch-day, precomputed remaining days, as-of day, and optional
    localized target label.
  - android widget: passive Glance `CountdownWidget` registered in the manifest,
    rendering only the snapshot written by the future sync layer rather than
    doing date/calendar computation inside the widget.
  - localization/tests: added fa/en countdown strings/plurals, widget picker
    metadata/preview, and `CountdownTest` for round-trips, state mapping, and
    malformed payload handling.
  **Next session:** build the Android date editor + sync writer so shared dates
  are uploaded and refreshed daily with locale/Jalali-aware labels.

- **2026-06-30** — Session 7: Added the Android "Their World" widget vertical
  slice and audited the five recently added widgets:
  - backend: added `their_world` as a valid generic widget kind so future sync
    can use `/v1/widgets/their_world` without schema changes.
  - android domain: `TheirWorld`, `WeatherCondition`, and `DayNightState` with a
    bounded wire format for partner location label, local time, UTC offset,
    weather, temperature, day/night state, and observation time.
  - android widget: passive Glance `TheirWorldWidget` registered in the manifest,
    showing partner time, weather, and day/night from a precomputed snapshot.
  - verification: re-parsed Android XML resources, checked layout text binding,
    ran git whitespace checks, and reviewed Love Tap, Shared Drawing, Shared
    Photo, Countdown, and Their World for passive rendering, bilingual
    resources, widget metadata, and backend kind compatibility; no build
    execution was possible in this environment.
  **Next session:** implement Android sync/data layer foundations so Mood, Love
  Tap, Drawing, Photo, Countdown, and Their World can start moving real payloads.
