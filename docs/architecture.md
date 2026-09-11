# Architecture

## Goals

Battle Bunny Wealth should ship as a simple deployable artifact while preserving hard boundaries between the game, seasonal economy, player identity, and experimental Proof-of-Play blockchain protocol.

The first production shape is a single Go binary that embeds the Vite SPA. Native mobile clients can be added later for platform attestation while sharing the same backend/protocol contracts.

## Product architecture

Battle Bunny Wealth contains three different competitions that share one player identity but must remain separate in code and economics:

```text
1. ECONOMIC GAME
   idle businesses -> seasonal Bunny Bucks -> season rewards

2. NETWORK GAME
   optional missions -> Proof-of-Play authority -> committee eligibility -> CARROT

3. SKILL GAME
   Warren Wars -> equal-start arena -> wins/rating/championships
```

A value from one system must not silently become authority or combat power in another.

Examples:

- Bunny Bucks do not become validator weight.
- CARROT does not buy ranked combat statistics.
- Proof-of-Play authority does not grant stronger bombs.
- Warren Wars wins do not automatically control consensus.

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
|    +-- player/profile services (future)                 |
|    +-- season/business services (future)                |
|    +-- Warren Wars services (future)                    |
|    +-- Proof-of-Play API                                |
|    +-- SPA fallback                                     |
|                                                         |
|  internal/proofofplay                                   |
|    +-- participant / device identity                    |
|    +-- challenge / mission / proof model                |
|    +-- authority state                                  |
|    +-- block + hash chain                               |
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

As the game grows, game-state packages should remain separate from `internal/proofofplay`. Warren Wars simulation should be deterministic and isolated enough to test without rendering or blockchain state.

## Frontend

The frontend uses SolidJS 2, `@solidjs/web`, Solid Router v2 prerelease, Vite, TypeScript, and StyleX. It is a client-side SPA. During local development Vite proxies `/api` to Go. Production builds into `web/dist`, which the Go compiler embeds.

StyleX is intentionally compile-time. Components define colocated atomic styles; a small global stylesheet is reserved for reset/document-level rules.

The SPA should expose the fiction, not protocol jargon. For example, a protocol challenge can appear as an NPC-issued mission rather than an epoch/attestation form.

## Backend

Use the Go standard library first. The server is deliberately `net/http` rather than a framework. Dependencies should be added only when the standard library cannot provide a clear implementation.

API prefix: `/api/v1`.

Current endpoints:

- `GET /api/v1/healthz`
- `GET /api/v1/proof-of-play`
- `GET /api/v1/proof-of-play/blocks/head`

Future API domains should include explicit boundaries for:

- player profile/customization
- NPC/story state
- seasons/businesses
- Warren Wars matchmaking/simulation/results
- Proof of Play missions/authority
- CARROT/network state

## Player and NPC boundary

The player creates and names their own bunny. Player profile state belongs to the account/game domain.

First Sergeant Hard-as-Nails, Da Champ, Private Stuffy, Corporal Boomboom, Captain Cashmere, Doc Flopsy, and future named characters are NPC/story content. Their identities are canonical world content and should not be confused with a player's protocol identity or ranked fighter statistics.

## Data/storage plan

v0 uses in-memory protocol scaffolding. Persistence is intentionally deferred until the core data model settles.

When persistence arrives, maintain explicit stores/boundaries:

```text
Game/account state
  - player-created bunny profile
  - cosmetics/collections
  - NPC/story progress

Season state
  - businesses
  - Bunny Bucks
  - season standings
  - turn-in history

Warren Wars state
  - deterministic match inputs
  - ratings/results
  - no wallet-derived combat stats

Proof-of-Play state
  - devices/attestations
  - missions/proofs
  - authority
  - committees/finality
  - canonical chain/state

CARROT ledger
  - fixed supply
  - deterministic issuance/allocation
```

The canonical Proof-of-Play/CARROT state must never trust client-reported game balances.

Likely first server persistence: SQLite/libSQL for account/game/season state plus an append-oriented store for protocol blocks/state snapshots. The protocol must define deterministic serialization before network replication is treated as canonical.

## Warren Wars architecture rule

Ranked match simulation must accept a standardized starting state independent of account wealth.

Conceptually:

```text
MatchSeed + Map + Players + Ruleset
                  |
                  v
        deterministic simulation
                  |
                  v
              MatchResult
```

Player cosmetics may affect presentation but must not alter the deterministic competitive state.

## Proof-of-Play boundary

Proof-of-Play authority is derived from qualified network missions and history. It is an input to committee selection, not to game combat.

The protocol layer should eventually own deterministic versions of:

- mission/challenge qualification
- authority accrual/decay
- device weighting
- committee selection
- votes/finality
- CARROT issuance

The game layer can render those concepts as story but cannot be their source of truth.

## Native/mobile plan

The web UI remains useful for marketing/admin/testing. A future iOS/Android client is required to use App Attest / Play Integrity correctly. Native clients should talk to the same Go API/protocol endpoints and hold device private keys in hardware-backed key storage when available.

ATProto remains a candidate for portable account/social identity. It does not replace device attestation or Proof-of-Play Sybil resistance.

## Deployment

The deployable server is one Go binary. This keeps container and Fly/Cloudflare-style deployments small and makes static asset/version skew impossible: backend and frontend ship together.
