# Attestation provider status

| Provider | Native spike | Server adapter | Default executable verification |
| --- | --- | --- | --- |
| Apple App Attest | `native/ios/AppAttestClient.swift` | `AppleAppAttestVerifier` + validator interface | Fails closed until a concrete Apple validator is configured |
| Google Play Integrity | `native/android/AttestedDeviceClient.kt` | `GooglePlayIntegrityVerifier` + REST decoder | Works when package/certificate policy and a valid OAuth access token are configured |
| Development | local/server only | deterministic test adapter | Available only with `BBWEALTH_DEV_CONTROLS=1`; never testnet eligible |

This status table is intentionally explicit so a provider interface or native API call is not confused with completed production verification.
