# Account and portable identity

v0.5 gives Battle Bunny Wealth real account ownership without collapsing player identity, login credentials, social identity, or device identity into the same thing.

## Account ownership

A Battle Bunny account owns its persistent Warren/game state. Game saves now live under `data/players/<account-id>.json`; the first authenticated account can automatically claim the legacy `data/game-state.json` save so v0.4 progress is preserved.

Authentication state lives separately in `data/identity.json`.

## Passkeys

The first authentication mechanism is WebAuthn/passkeys. The server:

- generates short-lived registration/login challenges
- validates the browser origin and RP ID
- requires discoverable/resident credentials for usernameless sign-in
- currently accepts ES256/P-256 credentials with `none` attestation
- verifies assertion signatures server-side
- tracks authenticator signature counters and rejects counter rollback
- stores only credential public material; no account password exists
- issues random opaque sessions whose hashes are persisted server-side
- stores the session token in an HttpOnly, SameSite=Strict cookie

This implementation is an alpha security boundary, not an audited authentication library. Before public production use it needs dedicated WebAuthn interoperability testing and security review; expanding authenticator algorithms/attestation formats should use a mature audited implementation unless there is a compelling protocol reason not to.

## Local configuration

Development defaults are:

```text
RP ID: localhost
Origins:
  http://localhost:8080
  http://localhost:5173
```

Production must set values appropriate to the deployed host:

```bash
BBWEALTH_RP_ID=battlebunnywealth.com
BBWEALTH_ORIGINS=https://battlebunnywealth.com
```

Multiple origins are comma-separated.

## ATProto profile link

v0.5 can resolve a `did:plc` document through `plc.directory` and record:

- DID
- claimed ATProto handle
- PDS service endpoint

The stored status is deliberately:

```text
resolved-unverified
```

Resolution proves that the DID document exists. It does **not** prove that the currently signed-in Battle Bunny account controls that DID. Therefore an ATProto link:

- is presentation/profile metadata only
- grants zero Proof-of-Play authority
- is not accepted as an authentication credential
- must not count toward one-human/one-account assumptions

A future ATProto OAuth/proof-of-control flow can promote a link to a verified status. `did:web` remains part of the architecture but is not resolved by the v0.5 runtime yet.

## Device enrollment

An account may enroll multiple device records. The v0.5 browser flow generates a P-256 device key, keeps the private `CryptoKey` locally in IndexedDB, and sends only its SPKI public key to the server.

Device records support explicit revocation.

Every v0.5 browser device is marked:

```text
attestationStatus: unattested
```

That means it is **not eligible for Proof of Play authority**. v0.7 adds the Apple App Attest / Google Play Integrity work needed to turn a device key into a stronger participation signal.

## Multi-device policy

Multiple devices are legitimate. Account storage therefore models a one-to-many account/device relationship instead of enforcing a fake one-phone rule.

This does not imply linear network influence. Proof-of-Play device weighting remains a later protocol decision and is expected to use diminishing weight plus participation history.

## HTTP API

v0.5 adds:

- `GET /api/v1/auth/me`
- `POST /api/v1/auth/passkey/register/begin`
- `POST /api/v1/auth/passkey/register/finish`
- `POST /api/v1/auth/passkey/login/begin`
- `POST /api/v1/auth/passkey/login/finish`
- `POST /api/v1/auth/logout`
- `PUT /api/v1/account/atproto`
- `DELETE /api/v1/account/atproto`
- `POST /api/v1/account/devices`
- `POST /api/v1/account/devices/{id}/revoke`

All `/api/v1/game/*` state is now account-scoped and requires an authenticated session.
