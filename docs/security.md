# Security and threat model

Proof of Play is fundamentally a Sybil-resistance experiment. Security design must assume attackers are financially motivated and automate everything that can be automated.

The game must also assume that anything tied to ranked Warren Wars, CARROT, season standings, or network authority will be attacked if it gains value.

## Threats

### Account farms

An attacker creates many DIDs/accounts. Mitigation: account identity alone grants almost no consensus weight; qualified attested devices, time, authority, rate limits, mission validity, and committee randomness matter.

### Device farms

An attacker operates racks of real phones. Hardware attestation cannot prevent this. Mitigation candidates: diminishing multi-device weight, slow authority growth, challenge diversity, network/device activity analysis, cost modeling, and caps.

### Mission farms / botting

An attacker automates Proof-of-Play missions to accumulate authority.

Mitigations:

- missions must correspond to independently verifiable protocol work
- unpredictable challenge inputs
- challenge expiry/replay protection
- rate limits and diminishing returns
- authority caps/bounded selection functions
- device attestation where applicable
- behavioral/network anomaly analysis as a secondary signal
- occasional human-liveness interactions only when they materially raise bot cost

Repetitive clicking alone must never create meaningful authority.

### Authority grinding

A highly active participant attempts to accumulate permanent dominant validator weight.

Mitigations:

- bounded authority
- slow accrual
- inactivity decay
- randomized committee selection
- recent-activity windows where appropriate
- no guarantee of selection at maximum authority
- simulation of capture probability before testnet

The goal is "active participants have better odds than inactive participants," not "the most active player becomes a permanent validator king."

### Simulator/emulator farms

Attestation provider policy should reject unsupported/untrusted environments for consensus eligibility. Development providers must never be accepted by production networks.

Simulator detection alone is insufficient; production eligibility should be based on cryptographically verified provider evidence and protocol policy.

### Patched clients

Never trust game-client state. Proofs are signed over server/network challenges and independently verified. Gameplay rewards and consensus eligibility need server-authoritative validation where abuse matters.

### Seasonal economy cheating

A client claims impossible Bunny Bucks/business progression to gain season standing or rewards.

Mitigations:

- authoritative server state for competitive/valuable season outcomes
- deterministic or auditable offline-earnings calculation
- validation of business upgrades and timestamps
- anti-replay/idempotency for economic actions

Seasonal Bunny Bucks must never be trusted as evidence of Proof-of-Play authority or CARROT entitlement.

### Ranked Warren Wars cheating

A modified client grants speed, extra bombs, altered collision, visibility advantages, or other combat power.

Mitigations:

- deterministic/server-authoritative simulation
- standardized starting state
- validate every gameplay-relevant input
- cosmetics excluded from simulation state
- replay/audit support for suspicious matches

Wallet balances, season wealth, CARROT, and Proof-of-Play authority must never alter ranked starting statistics.

### Replay

Challenges include unique IDs/nonces, epochs, expiry, and single-use tracking. A proof for one challenge must not validate for another.

### Key theft

Private device keys should be non-exportable/hardware-backed where possible. Account recovery and device revocation must not silently transfer old device authority to a new key.

### Attestation-provider outage/ban

The protocol cannot let one commercial provider permanently halt finality. Provider diversity, grace windows, and emergency governance/upgrades need explicit design before testnet.

### Colluding committee

Committee selection and quorum must be simulated under correlated identities/devices. Authority cannot become an unbounded rich-get-richer mechanism.

### Founder/admin allocation risk

The planned 20% founder/admin CARROT allocation creates concentration risk if unlocked or movable all at once.

Before economic activation:

- allocation must be public and deterministic
- genesis/account addresses must be documented
- vesting/lock policy should be defined
- circulating-supply reporting must distinguish locked vs. circulating founder supply
- founder holdings must not grant ranked combat advantages

### CARROT economic attacks

Before CARROT becomes economically valuable, model:

- wash trading
- reward farming
- committee bribery
- fee manipulation
- treasury compromise
- key loss
- market manipulation
- validator collusion
- Sybil attacks driven by token value

## What attestation does not prove

- one device = one human
- one account = one human
- the human is actively looking at the screen
- the device owner is unique
- an Apple/Google verdict can never be forged or misissued

Treat attestation as one costly-to-fake signal among several.

## Competitive fairness invariants

For ranked Warren Wars:

- every player starts from the same ruleset-defined combat capabilities
- persistent purchases are cosmetic only
- CARROT cannot buy combat power
- Bunny Bucks cannot buy combat power
- Proof-of-Play authority cannot buy combat power
- NPC/story progress cannot buy combat power

Any future feature that violates these invariants requires an explicit ADR and should default to a non-ranked/casual ruleset instead.

## Protocol code rules

- deterministic encoding before distributed consensus
- explicit version numbers on signed/protocol objects
- domain validation before cryptographic verification
- reject unknown attestation providers by default
- no development/test keys on public networks
- fuzz block/proof parsers before accepting untrusted network traffic
- deterministic authority calculation
- deterministic CARROT supply/issuance accounting
- no client-reported game balance used directly in consensus
