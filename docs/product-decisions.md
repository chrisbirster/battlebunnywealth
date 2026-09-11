# Product decisions

This page is the compact source of truth for decisions already made in Battle Bunny Wealth. Detailed rationale and implementation notes live in the linked design documents.

## Product identity

- Domain/product: **Battle Bunny Wealth** (`battlebunnywealth.com`).
- Tone: cartoon military parody, absurd wealth-building, and arcade bomb combat.
- The game should be fun first; the blockchain should not be the marketing prerequisite for understanding the game.

## Player and NPC roles

### Decided

- Every player creates and names their **own Battle Bunny**.
- Player customization is a major identity/progression system.
- Named characters such as First Sergeant Hard-as-Nails, Da Champ, Private Stuffy, Corporal Boomboom, Captain Cashmere, and Doc Flopsy are **NPCs**.
- NPCs provide story, lore, tutorials, jokes, rivalries, and missions.
- NPC identities do not replace the player's avatar and do not grant ranked combat statistics.

Principle:

> NPCs belong to the world. Your bunny belongs to you.

## The three core systems

### 1. Seasonal idle economy

- Players create and upgrade bunny businesses.
- Businesses generate seasonal game money: **Bunny Bucks**.
- Players try to maximize business income over a season.
- Season wealth can be turned in/spent for persistent collection, cosmetic, profile, story, trophy, and warren rewards.
- Bunny Bucks are not CARROT and do not have a fixed supply.

### 2. Proof of Play network game

- Optional missions are the main game-facing interface to useful network participation.
- Missions are not required for ordinary idle progression or Warren Wars access.
- Qualified missions increase **Proof-of-Play authority**.
- More active, honest participation gives a player a higher chance of being selected for network verification than an inactive participant.
- Higher authority is a probability/eligibility advantage, not guaranteed permanent validator status.
- Authority should grow slowly, be bounded, and decay gradually with inactivity.
- Multiple legitimate devices are allowed, but additional devices should have diminishing weight.

### 3. Warren Wars skill game

- Warren Wars is an original grid-bomb arena inspired by classic maze bomb games.
- Every ranked match starts competitively equal.
- Ranked outcomes are determined by movement, timing, prediction, trapping, map control, and adaptation.
- Persistent wealth, CARROT, purchases, account age, NPC/story progression, and Proof-of-Play authority cannot buy ranked combat power.
- Gameplay-affecting powerups are obtained inside the match under the same rules for all players.
- Persistent rewards can change appearance and status, not ranked statistics.

Principle:

> You can buy style. You can earn status. You cannot buy skill.

## Identity and attestation

### Decided direction

- Player profile, account identity, authentication identity, device identity, and Proof-of-Play authority are separate concepts.
- ATProto is a candidate for portable account/social identity and DIDs.
- ATProto/DIDs are **not** proof that an account represents one unique person.
- Passkeys are the preferred account-authentication direction.
- Apple App Attest is the primary iOS attestation candidate.
- Google Play Integrity plus hardware-backed signing is the primary Android candidate.
- IMEI/permanent hardware identifiers are not the protocol identity model.
- Attestation is one costly-to-fake signal, not proof of unique humanity.

## Proof of Play

### Core thesis

Instead of securing eligibility primarily by burning computation or owning stake, Proof of Play researches whether long-lived, attested, useful participation by ordinary users/devices can provide meaningful Sybil resistance and validator selection.

Conceptually:

```text
participant identity
+ attested device key
+ unpredictable challenges
+ qualified missions
+ participation history
+ bounded authority
+ random committee selection
= Proof-of-Play verification eligibility
```

### Important non-claims

- Button presses alone do not secure the chain.
- One phone does not prove one human.
- One DID does not prove one human.
- App Attest/Play Integrity do not eliminate real-device farms.
- Proof of Play is not yet a production-proven consensus mechanism.
- Security assumptions must be tested through simulation, permissioned testnets, and adversarial public testnets before valuable economic activation.

## CARROT

### Decided

- The planned network coin is named **CARROT**.
- CARROT will have a **fixed maximum supply**, conceptually following Bitcoin's scarcity principle rather than unlimited issuance.
- A declining/halving-like issuance model is a candidate and should be formally specified before activation.
- Proof-of-Play network participation is the primary candidate distribution mechanism for the network-reward portion of supply.
- **20% of total fixed CARROT supply is reserved for the founder/admin allocation.**
- This is currently intended as a supply allocation, not a permanent 20% tax on every transaction or reward.
- CARROT cannot buy ranked Warren Wars combat power.
- Bunny Bucks do not convert directly one-for-one into CARROT.

### Still open

- total maximum CARROT supply
- divisibility
- exact issuance/halving schedule
- exact percentage of the remaining 80% assigned to Proof-of-Play rewards, ecosystem/game rewards, community treasury, and liquidity/network bootstrapping
- founder/admin vesting or lock schedule
- fee model
- governance model
- genesis distribution mechanics
- economic activation date

These must be explicit and deterministic before an economically valuable public network launches.

## Seasons

### Decided

- Idle-business competition is seasonal.
- Business optimization and seasonal income are a distinct competition from Warren Wars and Proof of Play.
- Seasonal resets create a healthy boundary for large idle-game numbers.
- Season history, trophies, cosmetics, and collection/progression can persist.

### Still open

- season length
- exact reset rules
- exact turn-in conversion curves
- leaderboard/ranking rules
- how much story content advances per season

## Legal/product constraints

Before implementing real-money/token entry fees, pooled prizes, token sales, cash-equivalent tournament rewards, or similar mechanics, obtain dedicated legal/tax/app-store/product review.

The architecture should not assume that every technically possible CARROT mechanic is appropriate to ship.

## Development/release process

- Feature work happens on `feature/*` branches.
- Feature PRs merge into `dev`.
- Releases are prepared from `dev` into `main`.
- Version tags are created from released `main` commits.
- `main` is release-only.

See [releases.md](releases.md) for the detailed release process.
