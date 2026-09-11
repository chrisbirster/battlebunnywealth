# Battle Bunny Wealth docs

This directory is the source of truth for product, game, economy, and protocol decisions.

Start with [Product decisions](product-decisions.md) for the compact list of what is already settled versus what remains open.

- [Product decisions](product-decisions.md) — canonical decisions for player identity, NPCs, seasons, Warren Wars fairness, Proof of Play, and CARROT.
- [Architecture](architecture.md) — executable system boundaries, product domains, storage separation, and deployment model.
- [Game design](game-design.md) — player-created bunny, NPC/lore roles, seasonal idle businesses, missions, and equal-start Warren Wars.
- [Warren Wars](warren-wars.md) — deterministic arena rules, bombs, pickups, ranked fairness, replay format, inputs, bots, and v0.2 validation.
- [Proof of Play](proof-of-play.md) — authority, optional network missions, committee-selection research, and consensus phases.
- [Identity and attestation](identity-and-attestation.md) — player/account/device separation, ATProto, passkeys, Apple App Attest, and Google Play Integrity.
- [Economy](economy.md) — Bunny Bucks vs. CARROT, fixed-supply intent, founder/admin allocation, and ranked fairness boundaries.
- [Security](security.md) — Sybil resistance, mission/device farms, authority grinding, CARROT risks, anti-cheat, and threat model.
- [Development](development.md) — local workflow, repository layout, API conventions, and testing.
- [Releases](releases.md) — `feature/* -> dev -> main -> tag` lifecycle.
- [Roadmap](roadmap.md) — staged path from player/arena/season MVP through Proof-of-Play testnets and CARROT specification.
- [ADR 0001](adr/0001-proof-of-play-boundary.md) — Proof of Play is a separate protocol package.
- [ADR 0002](adr/0002-portable-identity-and-device-attestation.md) — portable account identity plus device attestation.

## Current non-negotiable design rules

1. Players create and name their own bunny; named Battle Bunny characters are NPCs for story/lore/missions.
2. Seasonal businesses generate Bunny Bucks; Bunny Bucks are not CARROT.
3. Network missions are optional for ordinary game progression but can increase bounded Proof-of-Play authority.
4. Active honest participants can receive better committee-selection odds than inactive participants; activity never guarantees permanent control.
5. Ranked Warren Wars begins from an equal competitive state.
6. CARROT, Bunny Bucks, purchases, authority, account age, and story progress cannot buy ranked combat power.
7. CARROT is planned as a fixed-supply network coin with a 20% founder/admin supply allocation; the remaining supply schedule and allocation still require formal specification.
8. Proof of Play remains a research protocol until its Sybil resistance and finality survive simulation and adversarial testing.
