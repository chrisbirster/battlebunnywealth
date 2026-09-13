# Product decisions

This page is the compact source of truth for decisions already made in Battle Bunny Wealth. Detailed rationale and implementation notes live in the linked design documents.

## Product identity

- Domain/product: **Battle Bunny Wealth** (`battlebunnywealth.com`).
- Tone: cartoon military parody, absurd wealth-building, and arcade bomb combat.
- The game should be fun first; the blockchain should not be the marketing prerequisite for understanding the game.

## Player and NPC roles

- Every player creates and names their own Battle Bunny.
- Player customization is a major identity/progression system.
- First Sergeant Hard-as-Nails, Da Champ, Private Stuffy, Corporal Boomboom, Captain Cashmere, Doc Flopsy, and other named characters are NPCs for story/lore/tutorials/missions.
- NPC identities do not replace the player's avatar and do not grant ranked combat statistics.

> NPCs belong to the world. Your bunny belongs to you.

## The three core systems

### Seasonal idle economy

Players create and upgrade businesses that generate seasonal **Bunny Bucks**. Season wealth can fund collection, cosmetic, profile, story, trophy, and Warren rewards. Bunny Bucks are not CARROT and have no fixed global supply.

### Proof of Play network game

Optional missions increase bounded Proof-of-Play authority. Authority improves validator-selection probability but does not guarantee permanent status. Multiple legitimate devices are allowed with diminishing contribution. Attestation and participation history are security signals, not proof of one-human-one-device.

### Warren Wars

Warren Wars is an original equal-start grid-bomb skill game. CARROT, Bunny Bucks, purchases, account age, season rank, NPC/story progress, and Proof-of-Play authority cannot buy ranked combat power.

> You can buy style. You can earn status. You cannot buy skill.

## Identity and attestation

Player profile, account identity, authentication identity, portable/social identity, device identity, and Proof-of-Play authority remain separate. Passkeys are preferred for account authentication. Apple App Attest and Google Play Integrity/hardware-backed keys are device-integrity inputs. Neither proves unique humanity.

## Proof of Play

Proof of Play researches whether long-lived, attested, useful participation can provide meaningful Sybil resistance and rotating validator selection without Proof of Work or stake ownership being the primary eligibility mechanism.

Important non-claims remain: button presses do not secure the chain; one phone/DID does not prove one human; device attestation does not stop a real-device farm; and the mechanism is not production-proven until it survives simulation and adversarial public testing.

## CARROT — decided in v0.10

- symbol: **CARROT**
- fixed maximum: **21,000,000**
- divisibility: **8 decimal places**
- founder/admin allocation: **20% / 4,200,000**
- Proof-of-Play issuance reserve: **60% / 12,600,000**
- ecosystem reserve: **10% / 2,100,000**
- community treasury: **5% / 1,050,000**
- security/public-goods reserve: **5% / 1,050,000**
- issuance uses 32 approximately four-year eras; each era receives half of the remaining participation reserve and the final era drains the exact remainder
- founder vesting uses a one-year cliff and four-year total linear vest from genesis
- missions build authority; they do not directly pay CARROT
- fees are supply-neutral and distributed to finality signers; v0.10 testnet minimum fee is zero
- founder custody target: 2-of-3; protocol treasuries: 3-of-5; seven-day key-rotation delay
- treasury spending is disabled in v0.10
- the testnet genesis commits the executable CARROT policy hash
- CARROT cannot buy ranked Warren Wars power
- Bunny Bucks do not convert one-for-one into CARROT

Still open before economic activation: public wallet/transaction design, concrete treasury key identities, governance mechanics, nonzero production fee parameters if any, legal/tax/app-store treatment, and whether/when economic activation is appropriate.

See [CARROT protocol](carrot-protocol.md).

## Seasons

Idle-business competition is seasonal. Business optimization/season income is distinct from Warren Wars and Proof of Play. Season history, trophies, cosmetics, and collection/progression can persist while seasonal economy values reset.

## Legal/product constraints

Before implementing real-money/token entry fees, pooled prizes, token sales, cash-equivalent tournament rewards, or similar mechanics, obtain dedicated legal/tax/app-store/product review.

## Development/release process

Feature work happens on `feature/*`, merges into `dev`, and releases are prepared from `dev` into release-only `main` with version tags from released commits.
