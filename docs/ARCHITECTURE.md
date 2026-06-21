# Kenâr — Architecture

Companion to [`ROADMAP.md`](../ROADMAP.md). This document explains *how* the
pieces fit; the ROADMAP owns *what* and *when*.

## 1. Principles

1. **Widgets are the product.** They are passive, stateless renderers. They
   never fetch or compute — they display whatever the sync layer last wrote into
   their Glance state. This respects battery and Android's widget IPC/bitmap
   limits.
2. **Server is a blind relay for content.** Widget payloads are
   client-side-encrypted (E2E). The server stores opaque blobs + non-sensitive
   metadata and routes them; it cannot read content.
3. **Premium is server-authoritative.** The app never trusts a local flag.
   Premium content is not delivered unless the server confirms an active,
   pair-level subscription (validated against Cafe Bazaar / Myket).
4. **Provider independence (Iran context).** SMS and Push are behind interfaces
   (`SMSProvider`, `PushProvider`) so Kavenegar/Pushe can be swapped. No
   Firebase / Google Play Services / foreign SaaS.
5. **Bilingual is a definition-of-done.** fa (default) + en everywhere; no
   hardcoded user-facing strings.

## 2. Backend (Go, hexagonal)

```
cmd/server                 process entrypoint, wiring, graceful shutdown
internal/
  config                   env-driven configuration (KENAR_*)
  domain/pair              pure entities + rules (User, Pair, Invite, codes)
  app/pairing              use cases orchestrating ports
  ports                    interfaces: repos, SMS, Push, Event pub/sub
  adapters/
    httpapi                HTTP transport (router, locale middleware)
    postgres   (TODO)      repository implementations
    redis      (TODO)      EventPublisher/Subscriber, presence
    pushe      (TODO)      PushProvider
    kavenegar  (TODO)      SMSProvider
    minio      (TODO)      object storage for photos/voice
  platform/{i18n,logger}   cross-cutting infrastructure
i18n/                      fa.json / en.json server message catalogs
```

**Dependency rule:** `domain` depends on nothing; `app` depends on `domain` +
`ports`; `adapters` depend on `ports`. The core never imports an adapter.

## 3. Real-time backbone

```
            A acts (e.g. sets mood)
                  │
                  ▼
        HTTP/WS write to backend
                  │
       ┌──────────┴───────────┐
       ▼                      ▼
  Postgres write        Redis PUBLISH pair:{id}
  (widget_state)              │
                             fan-out
                   ┌──────────┴───────────┐
        B socket connected?          B socket absent
                   │                       │
            WS push to B           Pushe wake (silent) → B fetches
                   │                       │
                   ▼                       ▼
        B writes Glance state and calls widget.update()
```

- **One generic event/state model** (`widget_state` table + `ports.Event`)
  backs every widget kind, so new widgets need no schema or pipeline change.
- **Debounce/coalesce** rapid updates (e.g. drawing strokes) before publish.
- **Images** are downscaled aggressively client-side before render; the widget
  receives a small bitmap/URL, never a full-resolution photo.

## 4. Android (Clean Architecture + MVVM/MVI)

```
presentation   Compose screens, ViewModels, UI state
domain         entities + use cases (e.g. Mood, pairing) — framework-free
data           repositories, DTOs, network (OkHttp/WS), DataStore, sync
widget         Glance widgets + receivers (passive renderers)
core           cross-cutting (locale switch, theme tokens, di)
```

- **Sync layer** owns the socket + Pushe wake handling and is the *only* writer
  of Glance widget state. Widgets read that state and render.
- **Locale:** AndroidX per-app locales (`autoStoreLocales`) drives the in-app
  language switch; layout direction (RTL/LTR) follows the active locale, so
  Compose layouts mirror automatically.

## 5. Open design tasks (tracked in ROADMAP)

- **E2E key exchange** between the two paired devices (the server must remain
  unable to read payloads). Likely an authenticated key agreement at pair time.
- **Presence** model in Redis (who is connected / viewing now) for Presence
  Pulse and Hold Hands.
- **Billing validation** flow against Cafe Bazaar / Myket receipts.
- **OEM background-kill** mitigation (MIUI/Samsung battery-whitelist guidance).
