<div align="center">

# کِنار · Kenâr

**Always beside you — همیشه کنارت**

A widget-first, presence-first private shared space for two.

</div>

---

Kenâr (Persian: کِنار — *"beside / next to"*) is **not a messenger** and **not a
social network**. Two people pair, and their shared space lives as Android
home-screen widgets on **both** phones. A change by one appears on the other's
widget in seconds — without either opening an app.

> See [`ROADMAP.md`](./ROADMAP.md) for the full vision, constraints, phased
> feature plan, and progress. **It is the single source of truth.**

## Highlights

- **Widget-first** with Jetpack Glance — the widgets are the product.
- **Fully bilingual** Persian (default) + English, full RTL/LTR.
- **Iran-context**: no Firebase / Google Play Services / foreign SaaS.
  Push via **Pushe**; SMS OTP via **Kavenegar**; self-hosted on one server.
- **Server-authoritative premium**, **E2E-encrypted** payloads (blind relay).

## Tech stack

| Layer    | Tech |
|----------|------|
| Android  | Kotlin, Jetpack Compose, **Jetpack Glance**, Coil, Hilt, MVVM/MVI |
| Backend  | **Go** (hexagonal), WebSocket + Redis pub/sub real-time |
| Data     | PostgreSQL, Redis, **MinIO** (S3-compatible) |
| Auth     | Phone OTP (Kavenegar, behind an interface) |
| Push     | Pushe (behind an interface) |
| Dist.    | Cafe Bazaar + Myket (not Google Play) |

## Repository layout

```
ROADMAP.md     Single source of truth (read first, every session)
docs/          Architecture & decision records
infra/         docker-compose (Postgres, Redis, MinIO), env template, db init
backend/       Go service — cmd/ internal/{config,domain,app,ports,adapters,platform}, i18n/
android/       Kotlin app — Compose UI + Glance widgets, fa/en resources
```

## Getting started

### 1. Infrastructure

```bash
cd infra
cp .env.example .env        # then edit the secrets
docker compose up -d        # Postgres, Redis, MinIO (+ backend)
```

Services: Postgres `:5432`, Redis `:6379`, MinIO `:9000` (console `:9001`).
The Postgres schema is bootstrapped from `infra/postgres/init/`.

### 2. Backend (Go)

```bash
cd backend
go test ./...               # run the unit tests
go run ./cmd/server         # serves on :8080 (KENAR_HTTP_ADDR)
```

Sanity-check the bilingual pipeline:

```bash
curl localhost:8080/healthz
curl "localhost:8080/v1/locale?lang=fa"   # Persian message
curl "localhost:8080/v1/locale?lang=en"   # English message
```

Configuration is via `KENAR_*` env vars — see `internal/config/config.go`.
No secrets in the repo.

> **Note:** the Go toolchain is not yet installed in the current dev
> environment; the backend is scaffolded and will be CI-compiled once Go is
> provisioned (ROADMAP §12). Code is stdlib-only at the foundation layer.

### 3. Android

Open `android/` in Android Studio (it generates the Gradle wrapper jar on
first sync), or from the CLI after `gradle wrapper`:

```bash
cd android
./gradlew :app:assembleDebug
./gradlew :app:testDebugUnitTest
```

Switch language in-app from the landing screen; Persian is the default and
the app otherwise follows the system locale.

## Real-time pipeline (the backbone)

```
A acts → backend writes (Postgres) + publishes (Redis)
       → if B's socket is connected: push over WebSocket
       → else: wake B via Pushe → B fetches state → updates the Glance widget
```

Target latency: a few seconds. See [`docs/ARCHITECTURE.md`](./docs/ARCHITECTURE.md).

## Contributing

Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`, `docs:`),
small atomic commits. A feature is **not done** until both `fa` and `en`
strings, layouts, and formatting are complete (bilingual definition-of-done).
