# Identity and attestation

## Separate the identities

Battle Bunny Wealth uses several identity concepts that must not collapse into one identifier:

1. **Player profile** — the player's self-created Battle Bunny: name, callsign, cosmetics, rank, Warren Wars record, season history, and social presentation.
2. **Account identity** — portable account identity/profile handle, potentially an ATProto DID.
3. **Authentication identity** — passkey/session credentials proving control of the account.
4. **Device identity** — a cryptographic keypair plus platform attestation proving that a recognized app/device environment created or controls the key.
5. **Proof-of-Play authority** — protocol reputation/eligibility derived from qualified network participation over time.

These concepts can be linked, but none should be treated as equivalent to another.

## Player-created bunny

The player's bunny belongs to the player. The player chooses its name/callsign and cosmetic identity.

Named characters such as First Sergeant Hard-as-Nails, Da Champ, Private Stuffy, Corporal Boomboom, Captain Cashmere, and Doc Flopsy are NPCs in the shared game world. They are not substitutes for player identity and should not be used as consensus identities.

A player profile can point to an account DID, but public game presentation should not require exposing device-attestation identifiers or raw protocol credentials.

## ATProto

ATProto is a candidate for profile/social portability and DID-based account identity. It can be useful for handles, social graph portability, public profile records, and cross-service identity.

ATProto does **not** provide proof-of-personhood and does not prevent someone from creating many accounts. Therefore:

```text
one DID != one human
one account != one validator
one profile != one device
```

DID count must never directly equal consensus weight.

Keep the protocol interface generic enough to support `did:plc`, `did:web`, or a non-ATProto identity later.

## Passkeys

Passkeys are the preferred authentication direction for proving control of an account without inventing another password system.

Passkeys authenticate account control. They do not prove unique humanity and do not replace device attestation for Proof-of-Play eligibility.

## Apple

Use App Attest as the primary iOS attestation candidate. The server verifies Apple-issued attestation/assertion material and binds a generated app key to the participant/device enrollment.

Do not use IMEI or assume a permanent hardware serial is available to ordinary apps.

## Android

Use Play Integrity as the primary Android attestation candidate, combined with a hardware-backed signing key when supported. Integrity verdicts are inputs to policy, not magical proof that one device equals one person.

## Enrollment sketch

```text
1. player creates/authenticates account
2. player creates Battle Bunny profile
3. account optionally binds portable DID/social identity
4. app generates device keypair
5. server sends enrollment challenge
6. app requests platform attestation bound to challenge/key
7. server verifies provider evidence
8. participant signs device binding
9. server records participant <-> public key <-> provider
10. qualified missions can begin contributing Proof-of-Play authority
```

## Multi-device behavior

A real player may legitimately own multiple devices. The protocol should support this rather than pretending one account always maps to one phone.

However, additional devices should not linearly multiply committee influence. Candidate device weighting therefore uses diminishing additional weight and must be simulation-tested.

## Privacy principles

- Never require IMEI/advertising IDs for consensus.
- Store provider evidence only as long as necessary.
- Prefer hashes/commitments where raw evidence is unnecessary.
- Make multi-device use legitimate rather than pretending it never happens.
- Do not leak social/profile data into public consensus state unless explicitly required.
- Do not expose device public keys as the player's public social identity unless necessary.
- Keep player cosmetics and story state outside canonical consensus unless a protocol use is explicitly designed.

## Provider dependence

Apple/Google are centralized trust inputs. Therefore attestation must sit behind a provider interface with protocol-level policy. Future providers could include TPM/WebAuthn attestation, OEM attestation, or other independently verifiable mechanisms.

Proof of Play should ultimately depend on a portfolio of costly-to-fake signals rather than treating any single platform verdict as absolute proof of a unique human.
