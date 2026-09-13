# Economy

Battle Bunny Wealth deliberately separates the seasonal game economy from the network economy.

```text
SEASONAL GAME ECONOMY
businesses -> Bunny Bucks -> season standings -> cosmetics/collection/progression

NETWORK ECONOMY
Proof of Play -> authority/eligibility -> committee finality -> CARROT release
```

## Bunny Bucks

Bunny Bucks are the seasonal idle-business currency. They can inflate, reset between seasons, and fund business upgrades or persistent cosmetic/status rewards. Bunny Bucks are not a blockchain asset by default and do not convert one-for-one into CARROT.

Seasonal wealth may fund cosmetics, uniforms, bomb/explosion skins, emotes, banners, titles, medals, trophies, Warren decorations, story/collection unlocks, and cosmetic arena themes.

It cannot purchase ranked Warren Wars power.

## Ranked Warren Wars rule

CARROT, Bunny Bucks, purchases, account age, season rank, cosmetics, and Proof-of-Play authority cannot grant extra starting bombs, larger starting blast radius, faster movement, extra health/lives, shorter fuse timers, stronger damage, privileged spawns, or equivalent ranked advantages.

> You can buy style. You can earn status. You cannot buy skill.

## CARROT — frozen v0.10 policy

CARROT is the planned native network coin for the Proof-of-Play layer. v0.10 specifies it for research/testnet accounting only; it is not economically activated.

- fixed maximum: **21,000,000 CARROT**
- decimals: **8**
- atomic unit: 100,000,000 atoms per CARROT
- founder/admin: **20% / 4,200,000 CARROT**
- Proof-of-Play issuance reserve: **60% / 12,600,000 CARROT**
- ecosystem reserve: **10% / 2,100,000 CARROT**
- community treasury: **5% / 1,050,000 CARROT**
- security/public goods: **5% / 1,050,000 CARROT**

All 21,000,000 are committed to named allocation/reserve accounts in the deterministic genesis accounting model. Network rewards transfer from the participation reserve, so there is no unlimited mint function.

See [CARROT protocol](carrot-protocol.md) for the executable schedule and invariants.

## Issuance

Proof-of-Play issuance uses 32 eras. Each era spans 210,240 finalized ten-minute-target epochs, approximately four years. Each era gets half of the remaining 60% participation reserve; the final era receives the exact remaining atomic units.

The first-era reward is approximately 29.96575343 CARROT per finalized height. The integer schedule is deterministic and exactly exhausts the 12,600,000-CARROT participation reserve.

Rewards are intended for unique finality signers, split equally after committee selection. Authority changes selection probability; it does not multiply the reward after selection. Ordinary missions build authority and do not directly pay CARROT.

## Founder vesting

The founder/admin allocation has a one-year cliff and four-year total linear vest from genesis, measured in finalized epochs. No founder amount is spendable before the cliff; at the one-year cliff, 25% is unlocked; the full allocation is unlocked at four years.

This is a fixed supply allocation, not a permanent 20% transaction/reward tax.

## Treasury and custody

Ecosystem, community, and security/public-goods reserves remain protocol-reserved in v0.10; treasury spending is disabled in the executable policy.

Future activation requires 3-of-5 treasury custody. Founder custody requires 2-of-3. Key rotation has a seven-day/1,008-epoch delay. Production key material is not committed to source control and no single application server should hold a signing quorum.

## Fees

v0.10 fee accounting is supply-neutral: fees are transferred to finality signers, not minted or burned. The research testnet minimum is zero atoms. A nonzero economic fee floor requires an explicit reviewed protocol upgrade.

## Candidate CARROT uses after review

Possible future uses include protocol fees, marketplace settlement, community-created content bonds, anti-spam deposits, cosmetics/collectibles, and carefully designed governance. CARROT is not required to enjoy the base idle game and cannot buy ranked combat power.

Any token sale, real-money entry fee, pooled prize, wagering-like mechanic, cash-equivalent reward, or other economically sensitive feature requires dedicated legal, tax, app-store, security, and product review before implementation.

## Non-goals

- Do not tokenize every inventory item.
- Do not make Bunny Bucks redeemable one-for-one for CARROT.
- Do not promise passive financial returns for leaving the app open.
- Do not let wallet size determine Warren Wars outcomes.
- Do not design the game around speculative token price.
