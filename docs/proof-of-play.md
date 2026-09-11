# Proof of Play

## Thesis

Proof of Play asks whether **long-lived, attested participation by ordinary users/devices** can become part of a blockchain's Sybil-resistance and validator-selection mechanism.

It is not "proof that somebody tapped a button." Taps are cheap to automate. The protocol combines multiple signals and costs:

```text
portable participant identity
        +
hardware-backed device key
        +
platform/device attestation
        +
unpredictable epoch challenges
        +
participation history / reputation
        +
rate limits + diminishing multi-device weight
        ->
validator eligibility weight
```

## Why a game

Most consensus clients expose protocol machinery directly. Battle Bunny Wealth gives those events a comprehensible fiction. An epoch challenge can become a squad job; a successful proof can advance an operation; validator reputation can map to service record/rank.

The game must remain fun if the protocol is turned off.

## Protocol objects

### Participant

A stable protocol identity. It can reference an ATProto DID but does not equate "one DID" with "one human."

### Device

A participant-bound public key whose private key should live in Secure Enclave / hardware-backed Android keystore where supported. Enrollment includes an attestation provider and evidence.

### Epoch

A bounded protocol period. The scaffold defaults to ten minutes. Epoch duration is a consensus parameter and must become versioned before testnet.

### Challenge

An unpredictable, short-lived challenge bound to an epoch. Challenges prevent precomputation and replay.

### Participation proof

An envelope containing participant/device/challenge/epoch references, attestation digest, and device signature. v0 implements envelope validation interfaces; real Apple/Google verification is a later adapter.

### Committee

A sampled set of eligible participants that proposes/attests blocks. The research target is stake-independent or stake-light selection based primarily on qualified participation weight.

### Block

The current scaffold implements a SHA-256 linked block structure. It is not yet a distributed consensus implementation.

## Candidate weighting

Initial research hypothesis:

- first qualified device: 100% base device weight
- second: 25%
- third: 10%
- fourth and later: 2% each
- reputation increases slowly and is capped
- inactivity decays reputation
- recently enrolled identities cannot immediately dominate committees

Those numbers are placeholders to simulate attacks, not final tokenomics.

## Human interactions

Do not require constant tapping. Target 1–4 meaningful, unpredictable interactions per day at most. Background protocol work should do most of the participation work.

Human interaction is a liveness/bot-cost signal, not a sole proof of personhood.

## Consensus research phases

### P0 — local protocol model

Deterministic block hashing, epochs, challenges, device/proof types, verification interfaces, threat model.

### P1 — attested participation

Native iOS/Android enrollment, App Attest/Play Integrity verification, hardware key challenge signing, replay prevention.

### P2 — simulator

Model honest users, multi-device households, phone farms, bot farms, colluding validators, offline rates, and attestation-provider outages. Tune committee and reputation rules from simulation rather than intuition.

### P3 — permissioned testnet

Multiple independently operated nodes, deterministic state transition, committee selection, voting/finality, networking, persistent storage, observability.

### P4 — adversarial public testnet

Open enrollment with no valuable token. Bug bounty, Sybil attacks, economic modeling, protocol upgrades.

### P5 — economic layer decision

Only after security data exists decide whether a transferable token is necessary at all.
