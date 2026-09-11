# Idle economy and player profile

v0.3 turns Warren Command from a fake counter into a persistent, server-authoritative idle economy.

## Player profile

The player creates their own Battle Bunny identity. The persistent profile currently stores:

- bunny name
- callsign
- fur style
- ear style
- uniform
- cosmetic/service rewards

Named characters such as First Sergeant Hard-as-Nails, Private Stuffy, Captain Cashmere, Corporal Boomboom, Da Champ, and Doc Flopsy remain NPCs used for story, tutorials, missions, and lore. They are not selectable player avatars.

Authentication/account ownership is intentionally deferred to v0.5. Until then the server hosts one local persistent player state.

## NPC orientation

v0.3 persists a six-step orientation sequence presented by First Sergeant Hard-as-Nails, Private Stuffy, Captain Cashmere, Corporal Boomboom, Da Champ, and Doc Flopsy. The sequence introduces businesses, offline income, the Warren Wars fairness boundary, and Season Zero without making those NPCs playable characters.

## Persistent state

The Go server owns game state. v0.3 uses a small versioned JSON persistence layer at `data/game-state.json` by default. Override it with `BBWEALTH_GAME_STATE`.

The file is written atomically through a temporary file and carries `schemaVersion: 1`. This keeps persistence dependency-free during the single-player alpha while preserving an explicit migration boundary for a later multi-user database.

Persisted state contains the profile, Bunny Bucks, business levels, current season progress, turn-in count, cosmetics, onboarding progress, and last accrual timestamp.

## Bunny Bucks

Bunny Bucks are seasonal game money. They are not CARROT and have no blockchain meaning.

The initial alpha wallet starts at 750 Bunny Bucks. Businesses earn automatically based on server time. A returning player receives at most eight hours of offline earnings for a single absence; time beyond that cap is intentionally forfeited.

The current businesses are:

| Business | Base income | First upgrade |
| --- | ---: | ---: |
| Questionable Carrot Logistics | 10/sec | 500 |
| Warren Scrap & Salvage | 4/sec | 350 |
| Strategic Tunnel Tolls | 25/sec | 1,200 |

Upgrade cost grows by 1.65x per level. Base income scales linearly with level.

## Synergies

v0.3 introduces the first cross-business bonuses:

- Carrot Logistics level 5 gives Scrap & Salvage +10% income.
- Scrap & Salvage level 5 gives Carrot Logistics +10% income.
- Strategic Tunnel Tolls level 5 gives every business +5% income.

These numbers are alpha tuning parameters rather than permanent economic promises.

## Season Zero turn-in

Season Zero has a prototype target of 25,000 earned Bunny Bucks. Reaching the target allows a manual turn-in.

A turn-in:

1. awards a cosmetic/service-record reward
2. resets Bunny Bucks to the alpha starting wallet
3. resets business levels to one
4. resets current season earnings
5. increments the persistent turn-in count

The first reward is the `Season Zero Pennant`; subsequent alpha turn-ins receive service stars.

Formal season scheduling, archived seasons, story chapters, and broader reset rules belong to v0.4.

## Ranked firewall

The idle economy is deliberately outside Warren Wars combat state. The game API reports `rankedPowerAffected: false`, and no idle field is consumed by the deterministic arena engine.

Bunny Bucks, business levels, season earnings, turn-ins, cosmetics, CARROT, Proof-of-Play authority, account age, and story progress cannot modify ranked starting stats.

## HTTP API

v0.3 adds:

- `GET /api/v1/game/state`
- `PUT /api/v1/game/profile`
- `POST /api/v1/game/businesses/{id}/upgrade`
- `POST /api/v1/game/onboarding/advance`
- `POST /api/v1/game/season/turn-in`
- `GET /api/v1/game/standings`

All mutations are validated by the Go service rather than trusted from browser state.
