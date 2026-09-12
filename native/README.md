# Native attestation spikes

`native/` contains source-level v0.7 integration spikes for the mobile security APIs that cannot be exercised by the Vite web client.

These files are reference implementations, not yet standalone Xcode/Gradle applications and not part of CI.

## iOS

`ios/AppAttestClient.swift` demonstrates:

- a Secure Enclave P-256 mission-signing key;
- X.509 SubjectPublicKeyInfo export compatible with the Go device-enrollment API;
- `DCAppAttestService` support detection;
- App Attest key generation;
- a server-issued, device-key-bound challenge hashed into `clientDataHash`;
- `attestKey` evidence submission;
- DER ECDSA mission signatures.

A real iOS host app still needs the Battle Bunny Wealth App ID, App Attest capability/entitlements, authenticated API transport, secure persistence/recovery policy, and integration testing on physical Apple hardware.

The Go server currently exposes an `AppleAttestationValidator` interface but does not ship a concrete Apple attestation-object/certificate validator. Until that validator is supplied, Apple completion intentionally fails closed.

## Android

`android/AttestedDeviceClient.kt` demonstrates:

- Android Keystore P-256 mission-key creation;
- StrongBox preference with fallback when StrongBox is unavailable;
- X.509 SPKI public-key enrollment;
- Play Integrity Standard request preparation;
- use of the server-issued key-bound `requestHash`;
- `SHA256withECDSA` mission signatures.

A real Android host app still needs the Play Integrity dependency, Play Console/Google Cloud configuration, application package/signing certificate configuration, authenticated API transport, and physical-device tests.

The server can call `decodeIntegrityToken` through `PlayIntegrityHTTPDecoder`. The current executable accepts a short-lived OAuth access token through `BBWEALTH_PLAY_INTEGRITY_ACCESS_TOKEN`; production should replace the static token source with normal service-account/ADC token acquisition and rotation.

## Security boundary

Attestation is an input to Proof of Play, not proof that one device equals one human. The v0.7 gate is intended for permissioned non-economic testnet research. CARROT is not issued and distributed consensus is not enabled by these spikes.
