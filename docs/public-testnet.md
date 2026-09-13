# v0.11 adversarial public testnet

v0.11 opens selected network surfaces for hostile testing while keeping the asset explicitly **TEST-CARROT with no economic value**.

## Role separation

Public full/observer nodes may announce themselves and relay/verify data. Running more full nodes never creates Proof-of-Play authority or committee votes.

Validator applications are open to submit, but submission is not activation. Applications prove possession of the proposed validator key. The application cannot self-assert authority or attestation status; a server-side eligibility verifier supplies those values. Accepted candidates still pass through the v0.9 30-day maturation and rate-limited activation registry.

TEST-CARROT wallet keys are independent of infrastructure node keys and participant/device validator keys.

## TEST-CARROT wallets and transactions

`carrot-wallet` creates an Ed25519 test wallet. Addresses use the `tcarrot1` prefix and derive from the public key. Private key files are stored with mode `0600`.

Transactions commit to network ID, sender, recipient, amount, fee, exact nonce, expiry height, and public key. Admission verifies sender derivation, signature, transaction ID, TTL, nonce, and bounded mempool capacity.

TEST-CARROT uses the frozen v0.10 CARROT ledger rules and conservation checks. Test funding in the harness comes from the Proof-of-Play participation reserve. This is test infrastructure, not a token sale or financially valuable asset.

A signed transaction can also be represented deterministically as a `test-carrot-transfer` testnet operation so block-operation hashing can carry the exact signed payload. v0.11 does not yet execute arbitrary public wallet transfers inside the replicated consensus state machine; that is the next hardening milestone.

## Public node discovery

Node announcements are self-signed by the node's Ed25519 infrastructure key and expire within 24 hours. The directory has total-entry and per-host limits. These limits constrain memory/connection amplification but do not claim Internet-scale DDoS resistance.

A farm of node keys still has zero committee votes because discovery identity is not validator identity.

## Governance and upgrades

v0.11 adds deterministic dry-run governance proposals and signed validator votes for protocol-upgrade and treasury-test proposals. Duplicate, unknown, or invalidly signed votes are rejected. Treasury proposals are permanently non-executing in this milestone because v0.10 keeps treasury spending disabled.

Protocol upgrade plans commit network ID, from/to version, activation height, CARROT policy hash, and minimum software version. Staging requires at least 1,008 blocks of notice, approximately seven days at the ten-minute epoch target.

## Adversarial gate

`go run ./cmd/pop-public-smoke -attackers 100 -gate` verifies bounded same-host node flooding, malformed transaction rejection, valid signed transaction admission, open validator-candidate queuing, and preservation of the first maturity-window activation cap.

## Deliberate limitations

v0.11 does not make CARROT economically valuable, enable treasury spending, establish exchange infrastructure, solve unique-human identity, prove public-Internet DDoS resistance, or make validator activation instantly permissionless. Public transfer execution in replicated state, peer-diversity tests across independent hosts, and finalized validator-set transition commitments remain later work.
