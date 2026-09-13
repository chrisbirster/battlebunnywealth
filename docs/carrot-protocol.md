# CARROT protocol specification — v0.10

v0.10 freezes the first deterministic CARROT supply/accounting rules for the permissioned research network. It does **not** economically activate CARROT, create a token sale, list a token, or make CARROT redeemable for money.

The executable policy lives in `internal/carrot`. Testnet genesis commits the SHA-256 policy hash so nodes cannot silently run different tokenomics.

## Fixed supply

- Symbol: `CARROT`
- Protocol version: `carrot/1`
- Maximum supply: **21,000,000 CARROT**
- Divisibility: **8 decimal places**
- Atomic unit: `1 CARROT = 100,000,000 atoms`
- Maximum atomic supply: `2,100,000,000,000,000`
- No operation may create supply above this amount.

All supply is committed to named allocation/reserve accounts at genesis. Network issuance transfers from the Proof-of-Play reserve; it is not an unbounded mint function.

## Allocation

| Bucket | Percent | CARROT |
| --- | ---: | ---: |
| Founder/admin | 20% | 4,200,000 |
| Proof-of-Play issuance reserve | 60% | 12,600,000 |
| Ecosystem reserve | 10% | 2,100,000 |
| Community treasury | 5% | 1,050,000 |
| Security/public-goods reserve | 5% | 1,050,000 |
| **Total** | **100%** | **21,000,000** |

There is no separate v0.10 liquidity allocation. Any future liquidity/bootstrap use must come from a disclosed treasury category through a reviewed protocol/governance process.

## Proof-of-Play issuance

The issuance reserve is released only as finalized heights advance.

- target epoch duration: 600 seconds (10 minutes)
- issuance era: 210,240 finalized epochs, approximately four years at the target cadence
- 32 eras
- each era receives half of the remaining participation reserve
- the final era receives the exact remaining atoms
- each era budget is divided deterministically over its 210,240 heights
- integer remainder atoms go to the earliest heights in the era
- after era 32, the scheduled participation reward is zero

The first-era reward is approximately `29.96575343 CARROT` per finalized height and declines by roughly half each era. The exact integer function is `carrot.RewardAtHeight`.

The 32-era construction exhausts exactly 12,600,000 CARROT without floating-point arithmetic or supply drift.

### Who receives issuance?

v0.10 defines the reward pool for the **finality signers** of a finalized committee decision. Selected committee members get one vote each; authority affects selection probability, not reward weight after selection. The reward pool is split equally among unique finality signers, with any indivisible remainder atoms assigned by sorted participant ID.

Ordinary Proof-of-Play missions do **not** directly pay CARROT. Missions build authority/eligibility. This avoids turning the game into a tap-to-earn mechanism.

v0.10 implements and tests this ledger transition package, but it does not yet expose public wallet transactions or economically valuable rewards. v0.11 can wire valueless test CARROT operations into the adversarial public testnet.

## Founder allocation and vesting

The founder/admin allocation is exactly 20% of maximum supply.

- allocation: 4,200,000 CARROT
- one-year cliff: 52,560 finalized epochs
- four-year total vesting: 210,240 finalized epochs from genesis
- before the cliff: 0 founder CARROT is spendable
- at the cliff: 25% is unlocked because vesting is linear from genesis
- after the cliff: unlock continues linearly by finalized height
- at/after height 210,240: the full founder allocation is unlocked

The executable ledger rejects founder transfers above the unlocked amount.

## Treasury reserves and custody

v0.10 reserves 20% for ecosystem/community/security purposes but **treasury spending is disabled in the executable v0.10 policy**. This prevents the research implementation from accidentally becoming an economic treasury system before governance and legal review.

Custody requirements are nevertheless frozen for future activation:

- founder custody: **2-of-3** keys
- protocol treasury custody: **3-of-5** keys
- key rotation: minimum **1,008 epochs** (seven days at the target cadence) before an approved replacement becomes active
- production key IDs/public keys are provisioned at network launch and are not hard-coded into source control
- no single application server may hold enough keys to satisfy a production threshold

Actual treasury-spend signatures/governance are intentionally not activated in v0.10.

## Fees

Fee accounting is supply-neutral.

- v0.10 testnet minimum fee: `0` atoms
- no fee burn in policy version 1
- a sender-paid fee is transferred, not minted
- fees are distributed equally among unique finality signers
- deterministic remainder atoms use sorted participant ID order

A non-zero production fee floor requires a reviewed protocol upgrade before economic activation. Changing the floor does not change the fixed-supply rule.

## Genesis commitment

The permissioned testnet protocol version is bumped to v2. `Genesis` contains `carrotPolicyHash`. `NewGenesis` computes it from `carrot.DefaultPolicy()`, and every node rejects a genesis whose policy hash differs from its executable CARROT policy.

This means a node cannot quietly change maximum supply, allocations, vesting, fees, or issuance schedule while claiming to be on the same network.

The v0.10 policy hash is frozen by CI:

`19433c11cd2e19e973c9f15a5168b2d9f9fe676e368cffc26d8a2d077cd7f496`

Any intentional policy change requires a protocol-version change, updated documentation, new test vectors, and a new genesis/network agreement.

## Supply reporting

`cmd/carrot-spec` exposes the executable specification and supply report.

```bash
go run ./cmd/carrot-spec -check
go run ./cmd/carrot-spec -height 1
go run ./cmd/carrot-spec -height 210241
```

The report separates maximum supply, founder unlocked/locked atoms, participation reserve released/remaining, treasury-reserved atoms, and modeled circulating supply.

`ValidateConservation` requires the sum of all ledger balances to remain exactly 21,000,000 CARROT.

## Ranked-game boundary

CARROT cannot buy ranked Warren Wars power. It cannot change starting speed, bombs, blast radius, health/lives, fuse behavior, spawn quality, or equivalent competitive statistics. Bunny Bucks also never convert one-for-one into CARROT.

## Economic activation gate

Before CARROT has monetary value, the project still requires at minimum:

- adversarial public testnet work in v0.11
- public Sybil/device-farm attempts
- wallet/transaction signature design and review
- treasury governance/signature implementation
- protocol upgrade process testing
- security review and bug bounty
- legal, tax, and app-store review
- explicit decision on whether/where economic activation is appropriate

v0.10 is a deterministic protocol specification and test implementation, not an investment product or promise of value.
