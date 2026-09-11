# Warren Wars

Warren Wars is the active competitive half of Battle Bunny Wealth. It is an original grid-based bomb arena designed around movement, timing, prediction, traps, destructible terrain, and temporary pickups.

## Ranked fairness contract

The ranked rule is non-negotiable: **every player starts every match with equal combat power**.

A v0.2 player starts with:

- one active bomb slot
- blast radius two
- base movement speed one
- no kick ability
- no remote detonator
- no persistent combat modifiers

The following may never increase ranked starting combat power:

- Bunny Bucks
- CARROT
- Proof-of-Play authority
- account age
- season wealth
- purchased cosmetics
- story progression
- NPC relationships

Persistent progression buys identity, status, cosmetics, lore, collection, and other non-power rewards. Combat advantages are earned only inside the current match and disappear when the match ends.

## v0.2 arena rules

The first arena is a deterministic 13 x 11 grid. A seed produces the same terrain, hidden pickups, and replay result every time.

Terrain:

- border and structural walls are indestructible
- crates are destructible
- floor is traversable
- spawn areas are kept clear
- seeded crate layout is mirrored so starting regions have equivalent structure

Players are solid for movement purposes. A player cannot move onto a tile currently occupied by another living player. Two players claiming the same empty destination in one simulation step are both blocked.

## Bombs

A normal bomb has a fuse and explodes orthogonally. Blast propagation:

1. begins at the bomb tile
2. travels up/down/left/right up to the bomb radius
3. stops at indestructible walls
4. destroys and stops at the first crate in a direction
5. triggers another bomb caught by the blast, creating a chain reaction
6. eliminates any bunny occupying an active explosion cell

The last surviving bunny wins. Simultaneous elimination can produce a draw. A match also draws when the configured maximum match time is reached.

## In-match pickups

Destroyed crates may reveal deterministic, seed-controlled pickups:

- **Bomb** — increases simultaneous active bomb capacity
- **Blast** — increases blast radius
- **Speed** — improves movement cadence
- **Kick** — allows the bunny to push a bomb one open tile
- **Remote** — newly placed bombs become remotely detonatable; the action detonates the owner's oldest remote bomb

All pickup effects are temporary and reset at the end of the match.

## Determinism and replay

The simulation does not depend on rendering or wall-clock time. State advances only through discrete ticks and explicit player inputs.

A replay contains:

- protocol version
- arena seed
- player specifications
- arena configuration
- one input frame per simulation tick

Running the same replay must result in the identical final simulation state. This property is required for future authoritative multiplayer, dispute analysis, spectator playback, and anti-cheat work.

## Inputs

The current web prototype supports:

- WASD or arrow keys — move
- Space — place bomb
- E — remote detonate after collecting Remote
- touch directional pad
- touch Bomb and Remote buttons

Input is an adapter around the simulation. The engine itself has no keyboard, touch, DOM, or rendering dependency.

## Bots

v0.2 includes deterministic local bots for testing. They:

- avoid immediately dangerous blast lines when possible
- move through valid floor cells
- place bombs near crates or opponents
- use remote bombs when available

They are test opponents, not final game AI.

## v0.2 validation

The automated arena test suite covers:

- identical ranked starting stats
- seeded mirrored map generation
- bomb/explosion and chain-reaction behavior
- player collision policy
- deterministic replay
- 1,000 seeded headless matches checked for deterministic state convergence and valid player bounds

Future networking must preserve this deterministic simulation rather than reimplementing combat rules separately on the server and client.
