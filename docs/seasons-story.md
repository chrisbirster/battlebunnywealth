# Seasons, story, and Warren progression

v0.4 turns the v0.3 Season Zero prototype into a formal server-timed campaign lifecycle.

## Season state machine

A season is one of:

```text
active -> turn-in -> archived -> next active season
```

The production alpha timings are seven days active plus a 24-hour turn-in window. The server owns the clock and transitions; clients do not decide when a season ends.

The first campaign sequence is:

1. Season 1: Welcome to the Warren — Broken Burrow
2. Season 2: Scrap Row — Scrap Row
3. Season 3: Carrot District — Carrot District

The catalog loops during the alpha so lifecycle behavior can be exercised repeatedly.

## Turn-in and archive

Bunny Bucks earned during the active window become the season score. Businesses stop producing once the turn-in phase begins. Each 25,000 unspent season earnings can be turned in for a cosmetic pennant/decorative service reward.

When the turn-in window expires, the server archives season identity, chapter/location, final earnings, turn-ins, completion time, trophy, and medal. The next season starts with 750 Bunny Bucks and level-one businesses. Archived results, awards, cosmetics, story progress, and Warren decorations persist.

## Story progression

NPCs remain shared story characters. They never become player avatars or ranked-stat packages. v0.4 adds persistent story beats featuring First Sergeant Hard-as-Nails, Private Stuffy, Captain Cashmere, Corporal Boomboom, Da Champ, and Doc Flopsy. Completing story beats changes narrative/service-record state, not Warren Wars power. Archiving a season advances the campaign chapter and unlocks the next location.

## Warren customization

The current themes are `field-camp`, `scrap-yard`, and `carrot-command`. Season pennants and trophies become Warren decorations. These are presentation/status rewards only.

## Economy telemetry

The server tracks lifetime Bunny Bucks earned, upgrades purchased, offline accrual count, offline earnings, and seasons completed for balancing. These values are development inputs, not player power.

## Accelerated development controls

Start locally with `BBWEALTH_DEV_CONTROLS=1`. Then `POST /api/v1/game/dev/season/advance` advances active -> turn-in or archives a turn-in season. Without the flag, the endpoint returns 404.

## Persistence migration

v0.4 raises game state to schema version 2. Existing v0.3 schema-1 saves migrate in place while preserving the existing player/economy data and initializing campaign/Warren/telemetry state.

## Ranked firewall

Season score, story progress, locations, Warren themes, trophies, medals, decorations, Bunny Bucks, CARROT, and Proof-of-Play authority remain outside deterministic Warren Wars combat state.
