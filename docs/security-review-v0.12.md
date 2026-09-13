# v0.12 external security review package

This document is the handoff index for an external review of the Battle Bunny Wealth Proof-of-Play public testnet as of protocol version 3.

The reviewed asset is TEST-CARROT only. It has no intended economic value. This package is not evidence that production CARROT is safe to launch.

## Review scope

Primary code:

- `internal/testnet/` — genesis, committee selection, proposal/vote validation, finality certificates, persistence, peer relay, recovery, activation policy, and state-machine boundary.
- `internal/publictestnet/` — public discovery, wallet transactions, mempool, validator admission, replicated TEST-CARROT state, finality settlement, network-transition commitments, upgrade planning, governance dry-runs, and public HTTP surface.
- `internal/carrot/` — fixed-supply policy, allocations, issuance schedule, founder vesting, fees, custody targets, and supply reporting.
- `cmd/pop-node/` — node assembly, persistence, synchronization, public HTTP wrapping, mempool-to-block wiring, and shutdown behavior.

Supporting evidence:

- `internal/popsim/`
- `cmd/pop-bootstrap-sim/`
- `cmd/pop-public-smoke/`
- `.github/workflows/ci.yml`

## Critical invariants to verify

1. Total CARROT atoms remain exactly `21,000,000 * 10^8`; transfers, fees, and rewards never create or destroy atoms.
2. Founder, participation, ecosystem, community, and security allocations remain fixed by the CARROT policy committed in genesis.
3. A TEST-CARROT transfer cannot execute without a valid signature, exact nonce, sufficient balance, correct network, valid expiry, and unique transaction ID.
4. Proposal state roots are derived by deterministic preview and the same transition is committed only after finality.
5. Restart and peer catch-up independently verify every finalized block and reproduce the same state root.
6. Authority affects committee-selection probability only. Once selected, one committee member has one vote.
7. Quorum is never reduced because validators are offline.
8. Node/observer count does not create consensus votes.
9. Finality rewards do not depend on a node-local quorum subset. The next finalized block canonically commits the prior vote set used for settlement.
10. Validator reward addresses are covered by the validator's signed application.
11. Public validator admission still requires server-side attestation/authority eligibility plus maturation/rate limits.
12. Protocol-upgrade and validator-set plans are commitments only; protocol v3 does not silently hot-swap executable consensus rules.
13. TEST-CARROT funding cannot mint new supply or spend treasury reserves through the public funding-policy path.
14. `main` contains no economically active production CARROT path as a result of v0.12.

## Adversarial cases to inspect

- forged proposal and vote signatures;
- duplicate votes and equivocation;
- vote replay across height, round, block, or network;
- stale/future transactions;
- nonce races and replacement attempts;
- duplicate transaction IDs;
- integer overflow around amount plus fee;
- invalid state roots;
- alternate valid quorum subsets for the same finalized block;
- malicious finality-settlement vote lists;
- corrupt/truncated `blocks.ndjson`;
- divergent genesis or CARROT policy hashes;
- stale-node catch-up from a malicious peer;
- oversized and multi-value JSON bodies;
- public peer floods from one host and one IP prefix;
- many observer nodes attempting to gain votes;
- real-phone validator farms;
- reward-address substitution during validator application forwarding;
- too-soon validator-set or protocol-upgrade commitments;
- malformed commitment hashes;
- provider outage and re-attestation edge cases;
- concurrent finalization/persistence failures.

## Commands reviewers should run

```bash
go test ./...
go vet ./...
go run ./cmd/pop-bootstrap-sim -phones 100 -gate
go run ./cmd/pop-cluster -blocks 20
go run ./cmd/carrot-spec -check
go run ./cmd/pop-public-smoke -attackers 100 -gate
```

Frontend/regression checks:

```bash
cd web
npm install
npm run test:arena
npm run typecheck
npm run build
```

## Files and protocol artifacts to compare

Reviewers should record:

- Git commit SHA;
- `testnet.ProtocolVersion`;
- genesis hash;
- `carrot.DefaultPolicy().Hash()`;
- `DefaultTestFundingPolicy().Hash`;
- validator-set commitment hash, if one is under review;
- protocol-upgrade commitment hash, if one is under review.

A finding against one of these hashes should not be assumed to apply to a different network revision without comparison.

## Known limitations accepted in v0.12

The following are known and should not be reported as hidden production guarantees:

- the consensus protocol is not formally verified;
- there is no full Tendermint-style prevote/precommit round-change implementation;
- adaptive corruption after committee reveal is not solved;
- no economic slashing exists;
- public discovery prefix limits do not provide ASN/geographic independence;
- validator-set commitments do not yet hot-activate a new committee set;
- protocol-upgrade commitments do not dynamically load or execute new consensus code;
- no BLS/threshold signature aggregation is used;
- no production treasury spending is enabled;
- the public funding endpoint exposes policy only, not an automatic faucet;
- Apple/Google attestation still introduces provider dependencies;
- device attestation is not proof of unique humanity;
- TEST-CARROT is not economically valuable and has not been tested under a real-money adversarial market.

## Release blockers for economically valuable CARROT

Before any economically valuable network launch, require at minimum:

- external consensus/cryptography/security review with critical/high findings resolved;
- independent wallet/key-management review;
- production validator-set activation and protocol-upgrade procedure review;
- long-running geographically/provider-diverse public testnet evidence;
- incident-response and key-compromise playbooks;
- treasury/custody operational review;
- legal review covering securities, money transmission, contests/gambling, tax, privacy, sanctions, and app-store rules as applicable;
- explicit production network genesis review and sign-off.

v0.12 passing CI is necessary evidence for continued research, not a production-safety certification.
