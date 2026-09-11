# ADR 0002: Portable identity plus device attestation

Status: Proposed

## Decision

Model account identity, authentication, and device identity separately. ATProto DID is the preferred first portable profile/account identifier. Passkeys authenticate account control. Apple App Attest and Google Play Integrity are preferred first device-attestation providers.

## Non-decision

A DID is not proof of personhood. A device is not a person. IMEI is not part of the protocol.

## Consequence

Consensus weight derives from qualified participation policy and history rather than raw account count.
