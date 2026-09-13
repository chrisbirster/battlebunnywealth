# Public-testnet security program

v0.11 is intended to be attacked in controlled, responsible ways.

In scope: signature/nonce bypasses, replay acceptance, supply-conservation failures, validator-activation bypasses, committee/finality safety failures, malformed-message crashes, peer-directory or mempool amplification, equivocation handling, attestation-binding failures, and upgrade/governance vote failures.

Out of scope: social engineering, attacks on Apple/Google/GitHub/cloud providers or unrelated third parties, destructive traffic against infrastructure you do not own, and attempts to create financial value from TEST-CARROT.

Reports should include reproduction steps, affected commit/version, expected versus actual behavior, and a minimal proof of concept. Do not publish exploitable details before maintainers have had a reasonable opportunity to patch them.

No monetary bug-bounty amount is promised by v0.11. A funded bounty can be announced separately if formal terms and budget exist.
