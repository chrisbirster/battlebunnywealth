# Security and threat model

Proof of Play is fundamentally a Sybil-resistance experiment. Security design must assume attackers are financially motivated and automate everything that can be automated.

## Threats

### Account farms

An attacker creates many DIDs/accounts. Mitigation: account identity alone grants almost no consensus weight; qualified attested devices, time, reputation, rate limits, and committee randomness matter.

### Device farms

An attacker operates racks of real phones. Hardware attestation cannot prevent this. Mitigation candidates: diminishing multi-device weight, slow reputation, challenge diversity, network/device activity analysis, cost modeling, and caps.

### Simulators/emulators

Attestation provider policy should reject unsupported/untrusted environments for consensus eligibility. Development providers must never be accepted by production networks.

### Patched clients

Never trust game-client state. Proofs are signed over server/network challenges and independently verified. Gameplay rewards and consensus eligibility need server-authoritative validation where abuse matters.

### Replay

Challenges include unique IDs/nonces, epochs, expiry, and single-use tracking. A proof for one challenge must not validate for another.

### Key theft

Private device keys should be non-exportable/hardware-backed where possible. Account recovery and device revocation must not silently transfer old device reputation to a new key.

### Attestation-provider outage/ban

The protocol cannot let one commercial provider permanently halt finality. Provider diversity, grace windows, and emergency governance/upgrades need explicit design before testnet.

### Colluding committee

Committee selection and quorum must be simulated under correlated identities/devices. Reputation cannot become an unbounded rich-get-richer mechanism.

## What attestation does not prove

- one device = one human
- one account = one human
- the human is actively looking at the screen
- the device owner is unique
- an Apple/Google verdict can never be forged or misissued

Treat attestation as one costly-to-fake signal among several.

## Protocol code rules

- deterministic encoding before distributed consensus
- explicit version numbers on signed/protocol objects
- domain validation before cryptographic verification
- reject unknown attestation providers by default
- no development/test keys on public networks
- fuzz block/proof parsers before accepting untrusted network traffic
