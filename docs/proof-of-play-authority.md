# Proof-of-Play missions and authority — v0.6

v0.6 turns the earlier Proof-of-Play design into an executable local prototype. It does **not** claim production consensus security.

## Player-facing rule

Network duty is optional.

A player can ignore Proof-of-Play missions and still use the seasonal idle economy, story progression, profile system, and equal-start Warren Wars. Completing qualified network missions only changes the player's network authority record.

## Mission flow

1. The player signs in with the v0.5 account.
2. The browser must have an active enrolled P-256 device key.
3. The player requests optional network duty.
4. The server creates a short-lived mission bound to:
   - account ID
   - device ID and enrolled public key
   - current Proof-of-Play epoch
   - current chain-head hash
   - unpredictable random challenge
   - assigned protocol operation
5. The browser signs the exact canonical mission payload with the local device private key.
6. The server verifies the signature against the enrolled public key.
7. A valid, non-replayed completion adds bounded prototype authority.

Current NPC-presented mission templates are:

- **Recon Patrol** — Private Stuffy / `verify-checkpoint`
- **Secure the Supply Line** — Captain Cashmere / `bind-device-checkpoint`
- **Verify Intel** — First Sergeant Hard-as-Nails / `sign-epoch-challenge`

Mission type assignment is deterministic from participant/device/epoch/ordinal inputs. The challenge itself is cryptographically random so it cannot be usefully precomputed.

## Rate limiting and replay defense

The current alpha constants are intentionally conservative and simulation-ready rather than final consensus values:

- mission TTL: **5 minutes**
- maximum missions issued per account per UTC day: **4**
- one still-valid pending mission per account/device is reused instead of issuing another
- completed mission IDs cannot be completed again
- expired missions cannot be completed
- a mission cannot be completed by another account, device ID, or public key

Issuance, including expired assignments, counts toward the daily ceiling. This prevents challenge-spam from becoming a way to grind mission selection.

## Authority formula

The initial executable authority policy is:

- valid mission: **+25 authority**
- maximum authority: **1,000**
- prototype committee eligibility threshold: **100**
- inactivity grace period: **72 hours**
- decay after grace: **50 basis points (0.5%) per full day**
- decay has a minimum one-point reduction per elapsed decay day while score is positive

These are protocol research parameters, not final tokenomics.

Authority is deliberately bounded. Completing more work after the cap does not create unlimited committee weight.

## New participant ramp

New participants do not immediately receive their full selection weight.

The prototype multiplier begins at 10% and linearly ramps to 100% across fourteen days from the participant's first valid mission. The effective committee weight is approximately:

```text
bounded authority
× newcomer multiplier
= prototype committee weight
```

A participant must also meet the 100-authority threshold before receiving any prototype committee weight.

This does not prove Sybil resistance; it simply makes instant account creation less useful and gives the simulator a concrete policy to attack in v0.8.

## Committee lottery

`AuthorityService.SelectCommittee` implements deterministic weighted sampling without replacement.

Inputs are:

- epoch
- external seed (currently suitable for a chain-head-derived prototype seed)
- target committee size
- all participant authority records

Eligible candidates are sorted deterministically. Each seat derives a SHA-256 draw from the seed, epoch, and seat number, then selects proportionally to the remaining bounded committee weights.

Higher authority therefore means **better odds**, not guaranteed selection.

This algorithm is a prototype. A production protocol still needs an unpredictable manipulation-resistant randomness construction, formal committee rules, replicated authority state, equivocation handling, and finality.

## Attestation gate

v0.6 proves possession of the **enrolled** device private key, but v0.5 browser device keys are still marked `unattested`.

Therefore every v0.6 authority response explicitly reports:

```text
prototypeEligible = true/false
productionEligible = false
```

and network status reports:

```text
productionCommitteeEnabled = false
```

v0.7 is responsible for replacing this missing trust signal with Apple App Attest / Google Play Integrity plus hardware-backed signing where available.

Until then, the committee implementation is for policy testing only.

## Persistence

Proof-of-Play mission/authority state is separate from account state and game economy state.

Default file:

```text
data/proof-of-play.json
```

Override with:

```bash
BBWEALTH_POP_STATE=/tmp/proof-of-play.json go run ./cmd/battlebunnywealth
```

Persisted data includes mission assignments, completion records, authority scores, first/last participation timestamps, and decay cursor state.

## Privacy boundary

Public mission responses do not include the internal account ID or the enrolled device public key. The completion history records a digest of the submitted signature rather than treating the raw signature as public identity data.

The committee selector currently operates inside the server process; v0.6 does not publish participant/account IDs as a public committee endpoint.

## What v0.6 proves

v0.6 demonstrates that the product can connect these concepts coherently:

```text
optional game-facing mission
        ->
short-lived chain-bound challenge
        ->
enrolled-device signature
        ->
replay-resistant completion
        ->
bounded persistent authority
        ->
weighted committee-selection input
```

It does **not** prove that one account equals one human, one device equals one human, an unattested browser device is genuine, or the resulting committee can safely secure economically valuable CARROT.
