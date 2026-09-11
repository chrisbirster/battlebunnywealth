# Identity and attestation

## Separate the identities

Proof of Play uses three different concepts that must not collapse into one identifier:

1. **Account identity** — portable identity/profile, potentially an ATProto DID.
2. **Authentication identity** — passkey/session credentials proving control of the account.
3. **Device identity** — a cryptographic keypair plus platform attestation proving that a recognized app/device environment created or controls the key.

## ATProto

ATProto is a candidate for profile/social portability and DID-based account identity. It does not provide proof-of-personhood and does not prevent someone from creating many accounts. Therefore DID count must never directly equal consensus weight.

Keep the protocol interface generic enough to support `did:plc`, `did:web`, or a non-ATProto identity later.

## Apple

Use App Attest as the primary iOS attestation candidate. The server verifies Apple-issued attestation/assertion material and binds a generated app key to the participant/device enrollment.

Do not use IMEI or assume a permanent hardware serial is available to ordinary apps.

## Android

Use Play Integrity as the primary Android attestation candidate, combined with a hardware-backed signing key when supported. Integrity verdicts are inputs to policy, not magical proof that one device equals one person.

## Enrollment sketch

```text
1. account creates/authenticates participant
2. app generates device keypair
3. server sends enrollment challenge
4. app requests platform attestation bound to challenge/key
5. server verifies provider evidence
6. participant signs device binding
7. server records participant <-> public key <-> provider
```

## Privacy principles

- Never require IMEI/advertising IDs for consensus.
- Store provider evidence only as long as necessary.
- Prefer hashes/commitments where raw evidence is unnecessary.
- Make multi-device use legitimate rather than pretending it never happens.
- Do not leak social/profile data into public consensus state unless explicitly required.

## Provider dependence

Apple/Google are centralized trust inputs. Therefore attestation must sit behind a provider interface with protocol-level policy. Future providers could include TPM/WebAuthn attestation, OEM attestation, or other independently verifiable mechanisms.
