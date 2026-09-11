# Architecture

## Goals

Battle Bunny Wealth should ship as a simple deployable artifact while preserving hard boundaries between the game, web UI, identity, and experimental blockchain protocol. The first production shape is a single Go binary that embeds the Vite SPA.

## Runtime

```text
                         +-----------------------+
                         |  SolidJS 2 SPA        |
                         |  Solid Router         |
                         |  StyleX               |
                         +-----------+-----------+
                                     |
                                /api/v1/*
                                     |
+-------------------- Go process ----v---------------------+
|                                                         |
|  net/http router                                        |
|    |                                                    |
|    +-- health/API                                       |
|    +-- game services (future)                           |
|    +-- Proof of Play API                                |
|    +-- SPA fallback                                     |
|                                                         |
|  internal/proofofplay                                   |
|    +-- participant / challenge / proof domain model     |
|    +-- block + hash chain                               |
|    +-- proof envelope validation                        |
|    +-- attestation verifier interface                   |
|    +-- signature verifier interface                     |
|    +-- committee/finality engine (future)               |
|                                                         |
|  web/dist embedded with go:embed                        |
+---------------------------------------------------------+
```

## Repository layout

```text
cmd/battlebunnywealth/   process entry point
internal/httpserver/     HTTP transport and middleware
internal/proofofplay/    protocol/domain code; no UI assumptions
web/                     Vite + SolidJS 2 SPA and go:embed package
docs/                    product/protocol source of truth
.github/workflows/       CI and tagged release builds
```

## Frontend

The frontend uses SolidJS 2, `@solidjs/web`, Solid Router v2 prerelease, Vite, TypeScript, and StyleX. It is a client-side SPA. During local development Vite proxies `/api` to Go. Production builds into `web/dist`, which the Go compiler embeds.

StyleX is intentionally compile-time. Components define colocated atomic styles; a small global stylesheet is reserved for reset/document-level rules.

## Backend

Use the Go standard library first. The server is deliberately `net/http` rather than a framework. Dependencies should be added only when the standard library cannot provide a clear implementation.

API prefix: `/api/v1`.

Current endpoints:

- `GET /api/v1/healthz`
- `GET /api/v1/proof-of-play`
- `GET /api/v1/proof-of-play/blocks/head`

## Data/storage plan

v0 is in-memory protocol scaffolding. Persistence is intentionally deferred until the core data model settles. When persistence arrives, keep canonical protocol state separate from game state:

```text
Game DB (accounts, progress, inventory)
              !=
Proof-of-Play canonical chain/state
```

Likely first server persistence: SQLite/libSQL for game state plus an append-oriented store for protocol blocks/state snapshots. The protocol must define deterministic serialization before network replication is treated as canonical.

## Native/mobile plan

The web UI remains useful for marketing/admin/testing. A future iOS/Android client is required to use App Attest / Play Integrity correctly. Native clients should talk to the same Go API/protocol endpoints and hold device private keys in hardware-backed key storage when available.

## Deployment

The deployable server is one Go binary. This keeps container and Fly/Cloudflare-style deployments small and makes static asset/version skew impossible: backend and frontend ship together.
