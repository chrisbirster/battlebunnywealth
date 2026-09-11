# Identity and attestation

## Separate the identities

Battle Bunny Wealth uses several identity concepts that must not collapse into one identifier:

1. **Player profile** — the player's self-created Battle Bunny: name, callsign, cosmetics, rank, Warren Wars record, season history, and social presentation.
2. **Account identity** — the persistent Battle Bunny account that owns game state.
3. **Authentication identity** — passkey/session credentials proving control of the account.
4. **Portable/social identity** — optional ATProto DID/profile metadata.
5. **Device identity** — a cryptographic keypair plus, eventually, platform attestation.
6. **Proof-of-Play authority** — protocol reputation/eligibility derived from qualified network participation over time.

These concepts can be linked, but none should be treated as equivalent to another.

## v0.5 implementation

v0.5 implements account ownership with passkeys. Each account gets its own persistent Warren/game-state file and can own multiple passkeys and multiple device records. The first account can claim the legacy single-player save.

The current WebAuthn verifier is intentionally narrow: ES256/P-256 discoverable credentials, `none` attestation, origin/RP/challenge validation, assertion signature verification, and signature-counter rollback detection. It is alpha authentication code and requires dedicated interoperability/security review before a public production launch.

## Player-created bunny

The player's bunny belongs to the player. The player chooses its name/callsign and cosmetic identity.

Named characters such as First Sergeant Hard-as-Nails, Da Champ, Private Stuffy, Corporal Boomboom, Captain Cashmere, and Doc Flopsy are NPCs in the shared game world. They are not substitutes for player identity and should not be used as consensus identities.

## ATProto

ATProto remains useful for profile/social portability and DID-based identity, but it does **not** provide proof-of-personhood and does not prevent someone from creating many accounts.

```text
one DID != one human
one account != one validator
one profile != one device
```

v0.5 resolves `did:plc` documents and stores the DID, handle, and PDS as an optional profile link with status `resolved-unverified`. Resolution alone does not prove that the signed-in Battle Bunny account controls the DID, so the link grants no authentication or Proof-of-Play authority. A later OAuth/proof-of-control flow is required before such a claim can be considered verified.

`did:web` remains supported by the architecture but is not resolved by the v0.5 runtime.

## Passkeys

Passkeys prove control of a Battle Bunny account without introducing passwords. They do not prove unique humanity and they do not replace device attestation for Proof-of-Play eligibility.

Session tokens are random opaque values; only their hashes are persisted server-side. Browser sessions use HttpOnly, SameSite=Strict cookies.

## Device enrollment

v0.5 adds a one-to-many account/device model with explicit revocation. The web client can generate a P-256 browser device key, persist the private `CryptoKey` in IndexedDB, and enroll only the public SPKI key with the server.

Every browser device in v0.5 is marked `unattested`. It is therefore not Proof-of-Play eligible.

## Apple

Use App Attest as the primary iOS attestation candidate. The server will verify Apple-issued attestation/assertion material and bind a generated app key to the participant/device enrollment.

Do not use IMEI or assume a permanent hardware serial is available to ordinary apps.

## Android

Use Play Integrity as the primary Android attestation candidate, combined with a hardware-backed signing key when supported. Integrity verdicts are inputs to policy, not magical proof that one device equals one person.

## Enrollment sketch

```text
1. player creates/authenticates account with passkey
2. account owns Battle Bunny game state
3. account may link portable/social identity
4. app generates a device keypair
5. v0.5 stores/revokes the public device identity
6. v0.7 adds platform attestation bound to the device key
7. qualified missions can later contribute Proof-of-Play authority
```

## Multi-device behavior

A real player may legitimately own multiple devices. The protocol supports this rather than pretending one account always maps to one phone.

Additional devices must not linearly multiply committee influence. Candidate device weighting therefore uses diminishing additional weight and must be simulation-tested.

## Privacy principles

- Never require IMEI/advertising IDs for consensus.
- Store provider evidence only as long as necessary.
- Prefer hashes/commitments where raw evidence is unnecessary.
- Make multi-device use legitimate rather than pretending it never happens.
- Do not leak social/profile data into public consensus state unless explicitly required.
- Do not expose device public keys as the player's public social identity unless necessary.
- Keep player cosmetics and story state outside canonical consensus unless a protocol use is explicitly designed.

## Provider dependence

Apple/Google are centralized trust inputs. Attestation must therefore sit behind a provider interface with protocol-level policy. Future providers could include TPM/WebAuthn attestation, OEM attestation, or other independently verifiable mechanisms.

Proof of Play should ultimately depend on a portfolio of costly-to-fake signals rather than treating any single platform verdict as absolute proof of a unique human.
