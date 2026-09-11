# Proof of Play

## Thesis

Proof of Play asks whether **long-lived, attested participation by ordinary users/devices** can become part of a blockchain's Sybil-resistance and validator-selection mechanism.

It is not "proof that somebody tapped a button" and it is not "the person who plays the most controls the chain." Taps are cheap to automate and raw playtime is easy to game.

The current design combines multiple signals and costs:

```text
portable participant identity
        +
hardware-backed device key
        +
platform/device attestation
        +
unpredictable epoch challenges
        +
completed optional network missions
        +
participation history / authority
        +
rate limits + diminishing multi-device weight
        ->
validator/committee eligibility weight
```

## Relationship to the game

Proof of Play is the network layer underneath Battle Bunny Wealth, but ordinary game progression must remain enjoyable without understanding or optimizing the protocol.

The game exposes network work through story missions. The player sees First Sergeant Hard-as-Nails assigning Recon Patrol, Checkpoint Duty, or Secure the Supply Line. The protocol sees a device answering unpredictable challenges, validating data, signing state, or participating in a committee.

The game must still function if Proof of Play is unavailable.

## Optional missions

Missions are the primary source of **Proof-of-Play authority**.

Important rule:

> Missions are optional for the idle economy and Warren Wars, but active and honest mission participation increases a player's probability of being selected for network verification duties.

A player who ignores network missions can still:

- create and customize a bunny
- build businesses
- generate seasonal Bunny Bucks
- progress through ordinary story content
- enter Warren Wars
- compete under the same ranked match rules

They simply accumulate less network authority than a player who consistently completes valid Proof-of-Play missions.

## Authority

Authority is a protocol reputation/eligibility signal. It is **not combat power, money, or voting power that grows without bound**.

Authority should reflect qualified participation over time rather than raw clicks.

Candidate behavior:

- valid missions add authority gradually
- repeated or low-value actions are rate limited
- recently created identities begin with little authority
- authority is capped or transformed through a bounded weighting function
- inactivity causes slow decay rather than instant loss
- invalid, equivocal, fraudulent, or abusive participation can reduce or suspend eligibility
- multiple devices for one person are legitimate but receive diminishing additional weight

This produces the intended relationship:

```text
inactive participant
lower committee-selection probability

occasionally active participant
moderate probability

consistently active + honest participant
higher probability
```

Higher authority means **more chance**, not guaranteed selection.

## Protocol objects

### Participant

A stable protocol identity. It can reference an ATProto DID but does not equate "one DID" with "one human."

### Device

A participant-bound public key whose private key should live in Secure Enclave / hardware-backed Android keystore where supported. Enrollment includes an attestation provider and evidence.

### Epoch

A bounded protocol period. The current scaffold defaults to ten minutes. Epoch duration is a consensus parameter and must become versioned before testnet.

### Challenge

An unpredictable, short-lived challenge bound to an epoch. Challenges prevent precomputation and replay.

### Mission

A game-facing assignment that may wrap one or more protocol tasks. Not every story mission has to affect consensus. Only missions that produce independently verifiable protocol evidence can contribute authority.

### Participation proof

An envelope containing participant/device/challenge/epoch references, attestation digest, and device signature. v0 implements envelope validation interfaces; real Apple/Google verification is a later adapter.

### Authority record

A deterministic protocol-state representation of the participant's qualified history, decay, penalties, and capped selection weight. The exact formula remains a research item and must be simulation-tested.

### Committee

A randomly sampled set of eligible participants that proposes and/or attests blocks. Selection should be unpredictable and weighted by bounded qualified authority rather than raw wealth or raw device count.

### Block

The current scaffold implements a SHA-256 linked block structure. It is not yet a distributed consensus implementation.

## Candidate weighting

Initial research hypothesis for device contribution:

- first qualified device: 100% base device weight
- second: 25%
- third: 10%
- fourth and later: 2% each
- authority increases slowly and is capped
- inactivity decays authority gradually
- recently enrolled identities cannot immediately dominate committees

Those numbers are placeholders for simulation and attack analysis, not final tokenomics or consensus constants.

## Committee selection sketch

Conceptually:

```text
all enrolled participants
        |
filter valid device/identity state
        |
calculate bounded authority weight
        |
use unpredictable protocol randomness / VRF-style selection
        |
small verification committee
        |
committee votes/attests
        |
quorum finalizes block
```

A future implementation must specify randomness, committee size, quorum, equivocation handling, liveness rules, and finality formally.

## Human interaction

Do not require constant tapping. The idle game should remain genuinely idle.

Human interaction can be an occasional liveness/bot-cost signal, but the useful network task matters more than repetitive screen interaction. A mission should not become a CAPTCHA job or punish players financially for not opening the app every few hours.

The exact mission cadence is TBD. It should be low enough that normal players can participate without compulsive engagement.

## CARROT relationship

CARROT is the planned fixed-supply network coin. Proof of Play is the primary candidate mechanism for distributing the network-participation portion of CARROT issuance.

A participant should not receive CARROT merely for having the game installed or for generating Bunny Bucks. Rewards should correspond to qualified protocol participation under deterministic issuance rules.

Authority and CARROT remain distinct:

```text
authority
= chance/eligibility to perform verification work

CARROT
= scarce network asset/reward
```

Large CARROT ownership must not directly purchase ranked Warren Wars power.

## What Proof of Play is trying to make costly

Every public blockchain needs a reason that one attacker cannot cheaply pretend to be millions of independent participants.

Proof of Play's research hypothesis is that an attacker should need some combination of:

- many legitimate attested devices
- many persistent identities
- meaningful elapsed time
- valid unpredictable challenge responses
- sustained participation history
- ongoing operational effort

rather than merely creating millions of accounts.

Whether those costs are sufficient to secure valuable consensus is an open research question, not an assumption.

## Consensus research phases

### P0 — local protocol model

Deterministic block hashing, epochs, challenges, device/proof types, verification interfaces, threat model.

### P1 — mission/authority model

Deterministic authority state, mission qualification, bounded weighting, inactivity decay, penalties, and selection inputs.

### P2 — attested participation

Native iOS/Android enrollment, App Attest/Play Integrity verification, hardware key challenge signing, replay prevention.

### P3 — simulator

Model honest users, inactive players, highly active players, multi-device households, phone farms, bot farms, colluding validators, offline rates, mission-completion distributions, and attestation-provider outages. Tune committee and authority rules from simulation rather than intuition.

### P4 — permissioned testnet

Multiple independently operated nodes, deterministic state transition, committee selection, voting/finality, networking, persistent storage, observability.

### P5 — adversarial public testnet

Open enrollment with valueless or non-transferable test assets first. Run bug bounties, Sybil attacks, economic modeling, and protocol-upgrade exercises.

### P6 — CARROT economic activation

Only after adversarial data exists should the fixed-supply CARROT issuance schedule and economically valuable network rewards activate.
