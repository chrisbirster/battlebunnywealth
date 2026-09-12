# Attested participation — v0.7

v0.7 adds the first provider-backed device-integrity boundary to Proof of Play. It does **not** make Battle Bunny Wealth a production blockchain and it does not prove unique humanity.

## Security model

The enrollment chain is:

```text
passkey account
    +
enrolled P-256 device/mission public key
    +
short-lived server challenge
    +
SHA-256 binding to that device public key
    +
platform provider evidence
    ->
attestation record
    ->
permissioned-testnet eligibility input
```

The canonical attestation binding is:

```text
bbw-attestation/v1
challenge=<random nonce>
device=<device id>
publicKeySha256=<SHA-256 of enrolled SPKI>
```

The server hashes that binding again into `requestHash`. Android sends this value through Play Integrity. iOS hashes the binding payload as App Attest `clientDataHash`.

Challenges expire after five minutes, are single-use, and are consumed before the external provider verification request so concurrent replays cannot race through the same nonce.

## Providers

### Apple App Attest

The provider boundary is `AppleAttestationValidator`.

The iOS spike generates an App Attest key and sends an attestation object tied to the server challenge. It also creates a Secure Enclave P-256 mission-signing key whose public SPKI is included in the server binding.

The repository deliberately does **not** claim full Apple server verification yet. `cmd/battlebunnywealth` currently constructs the Apple provider without a concrete `AppleAttestationValidator`, so App Attest completion fails closed until that validator is implemented and reviewed.

### Google Play Integrity

The Android spike uses Play Integrity Standard requests with the server-generated `requestHash` and uses Android Keystore for the mission-signing key, preferring StrongBox where available and explicitly falling back to the Android Keystore when StrongBox is unavailable.

The Go server includes `PlayIntegrityHTTPDecoder`, which calls Google `decodeIntegrityToken` and validates the returned policy signals:

- request hash matches the key-bound challenge;
- package name matches configuration;
- `PLAY_RECOGNIZED` is required;
- an optional application signing-certificate allowlist can be enforced;
- `MEETS_DEVICE_INTEGRITY` or `MEETS_STRONG_INTEGRITY` is required for a verified provider record;
- only `MEETS_STRONG_INTEGRITY` sets the conservative v0.7 `hardwareBacked`/permissioned-testnet eligibility signal;
- deployments may reject non-strong verdicts entirely with `BBWEALTH_ANDROID_REQUIRE_STRONG_INTEGRITY=1`.

That distinction is deliberate: a normal device-integrity verdict is useful evidence, but v0.7 does not overstate it as proof that the separate mission-signing key is hardware-non-exportable on every Android version/device configuration.

For the current executable, the Google decoder receives a short-lived OAuth bearer token through `BBWEALTH_PLAY_INTEGRITY_ACCESS_TOKEN`. Production deployment must replace this with normal service-account/ADC token acquisition and rotation.

### Development provider

The development provider exists only when `BBWEALTH_DEV_CONTROLS=1`. It accepts a deterministic test response and is permanently marked:

```text
hardwareBacked = false
productionEligible = false
```

It cannot satisfy the attested committee gate.

## Stored state

Attestation state is separate from game, identity, and authority state:

```text
data/attestation.json
```

Override with:

```text
BBWEALTH_ATTESTATION_STATE=/path/to/attestation.json
```

A successful provider check is also projected into the account device record so the account UI can display provider, verification state, hardware-backed signal, and permissioned-testnet eligibility.

## HTTP API

Authenticated endpoints:

```text
GET  /api/v1/proof-of-play/attestations
POST /api/v1/proof-of-play/attestations/begin
POST /api/v1/proof-of-play/attestations/complete
```

`begin` accepts:

```json
{"deviceId":"...","provider":"apple-app-attest"}
```

or:

```json
{"deviceId":"...","provider":"google-play-integrity"}
```

The provider must match the enrolled device platform. Native clients then submit provider evidence to `complete`.

## Multi-device weighting

Multiple devices are legitimate, but they do not linearly multiply authority. v0.7 applies the current research schedule to mission awards:

```text
first active device       100%
second active device       25%
third active device        10%
fourth and later            2%
```

With the current base mission award of 25 authority, the first device earns 25, the second 6 after integer rounding, the third 2, and later devices at least 1 when they complete a valid mission.

These values remain simulation parameters for v0.8, not frozen consensus constants.

## Committee gate

There are now two distinct concepts:

```text
prototypeEligible
= authority score satisfies the v0.6 threshold

productionEligible
= prototype threshold
  AND account has an active provider-verified hardware-backed device signal
```

The name `productionEligible` is inherited from the prototype model, but in v0.7 it means only **eligible for permissioned non-economic testnet committee research**. Distributed production consensus remains disabled in network status and is a later milestone.

## What attestation still does not prove

- one attested phone is not one unique human;
- an ATProto DID is not proof of personhood;
- a passkey is authentication, not personhood;
- platform providers are centralized trust inputs;
- a successful provider verdict does not make authority parameters economically secure;
- Play Integrity device-integrity evidence is not by itself proof that a separate mission key is physically non-exportable on every device configuration;
- provider outages and policy changes require explicit protocol behavior.

Those are reasons v0.8 is a simulator milestone rather than immediate token issuance.
