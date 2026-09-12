# v0.7 identity/attestation notes

v0.7 advances the device layer from an `unattested` browser-key prototype to a provider-backed enrollment boundary for native clients.

It does not change the core identity rule:

```text
passkey account != ATProto DID != device != human != Proof-of-Play authority
```

A provider verdict can make an active device eligible as one input to a permissioned non-economic testnet committee. It does not create authority by itself; the account must still satisfy the bounded mission-authority threshold.

See [Attested participation](attested-participation.md) for the key-bound challenge format, Apple/Google adapter status, native spikes, multi-device weighting, and known limitations.
