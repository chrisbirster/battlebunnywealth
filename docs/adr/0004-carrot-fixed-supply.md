# ADR 0004: CARROT uses genesis reserves and deterministic release

## Status

Accepted for v0.10 research/testnet use.

## Decision

CARROT has a fixed maximum supply of 21,000,000 with 8 decimal places. All supply is committed to fixed allocation/reserve accounts at genesis rather than created through an unbounded mint operation.

Allocation is 20% founder/admin, 60% Proof-of-Play issuance, 10% ecosystem, 5% community treasury, and 5% security/public goods.

Proof-of-Play issuance transfers from the participation reserve according to deterministic four-year eras. Each era receives half of the remaining reserve; the final era receives all remaining atoms. The testnet genesis commits the executable CARROT policy hash.

Founder allocation has a one-year cliff and four-year total linear vest. Treasury spending remains disabled in v0.10. Fee accounting is supply-neutral and does not burn or mint CARROT.

## Why

A reserve-at-genesis design makes the maximum-supply invariant simple to verify: accounting can move CARROT but cannot create more than the committed 21,000,000. A policy hash in genesis makes tokenomic rules part of network identity instead of operator configuration.

The declining release schedule preserves the intended scarce, Bitcoin-like issuance philosophy without copying Proof of Work or claiming equivalent security.

## Consequences

- changing tokenomics requires a protocol/genesis change rather than a server setting;
- all arithmetic is integer atomic-unit arithmetic;
- missions influence authority but do not directly become tap-to-earn payouts;
- treasury governance and economically valuable transfers remain intentionally deferred;
- old v0.9 permissioned-testnet genesis files are protocol-v1 artifacts and must not be treated as v0.10 genesis files.
