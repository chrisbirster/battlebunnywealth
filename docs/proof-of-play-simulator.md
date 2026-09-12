# Proof-of-Play simulator — v0.8

v0.8 turns the current Proof-of-Play authority and attestation rules into a deterministic Monte Carlo research tool.

The purpose is not to prove that Proof of Play is secure. The purpose is to make assumptions measurable and falsifiable before a valuable public network exists.

## What it models

The simulator starts from the executable v0.7 policy defaults:

```text
mission award                 25 authority
maximum authority          1,000
daily mission ceiling          4
eligibility threshold        100
inactivity grace               3 days
daily inactivity decay       0.5%
newcomer ramp                 14 days
committee size                64
quorum                       2/3
additional device weights 100 / 25 / 10 / 2 percent
```

Those values are imported from the current Go protocol configuration rather than maintained as a separate hand-written constants file.

Each simulated cohort can vary:

- account count;
- honest or adversarial behavior;
- first enrollment day;
- devices per account;
- fraction able to satisfy the attestation gate;
- daily activity probability;
- mission-completion probability;
- committee online probability;
- daily device churn probability;
- re-attestation/replacement delay;
- platform attestation provider;
- illustrative device/account/operating costs.

The simulator also models correlated Apple and Google provider availability.

## Built-in scenarios

```text
baseline
inactive-population
multi-device-households
unattested-bot-farm
phone-farm
delayed-phone-farm
provider-outage
device-churn
colluding-third
compromised-validators
smoke
```

`unattested-bot-farm` deliberately gives the attacker a very large number of cheap automated accounts that can build prototype mission history but cannot enter the attested committee set.

`phone-farm` instead gives the attacker one attested account per real device. This is the attack class Proof of Play ultimately has to make economically unattractive; attestation by itself does not eliminate it.

`delayed-phone-farm` starts the attacker halfway through the run so the newcomer ramp can be measured against an established honest population.

`provider-outage` applies correlated provider downtime. v0.8 uses a conservative **fail-closed** assumption: when a provider is unavailable, affected participants are temporarily excluded from qualified participation/committee selection for that draw. This is a research policy, not yet a finalized testnet liveness rule.

`device-churn` models phone replacement/re-enrollment delays.

`compromised-validators` models already-attested participants whose keys/behavior have become adversarial; their acquisition cost is intentionally different from buying new devices.

## Running it

Baseline JSON:

```bash
go run ./cmd/pop-sim -scenario baseline -seed 42 -format json
```

All built-in scenarios as CSV:

```bash
go run ./cmd/pop-sim -scenario all -seed 42 -format csv
```

Real-phone attacker population sweep:

```bash
go run ./cmd/pop-sim -scenario attack-sweep -seed 42 -format csv
```

Write a report to disk:

```bash
go run ./cmd/pop-sim \
  -scenario phone-farm \
  -seed 20260912 \
  -days 90 \
  -trials 10000 \
  -format json \
  -out /tmp/phone-farm.json
```

List scenario names:

```bash
go run ./cmd/pop-sim -list
```

Taskfile shortcuts:

```bash
task sim
task sim:all
task sim:sweep
```

## Determinism

A simulation is deterministic for the tuple:

```text
model version
+ scenario configuration
+ seed
```

The same seed and input must produce the same report. CI has an explicit test for this property.

Different seeds should be used when estimating uncertainty across independent runs.

## Authority model

Each simulated day:

1. staged cohorts become active when their `startDay` is reached;
2. authority decay is applied after the configured inactivity grace;
3. device churn may temporarily remove an attested participant while it replaces/re-attests the device;
4. a correlated provider availability result is drawn for Apple and Google;
5. active participants attempt up to the configured daily mission limit;
6. successful missions add authority according to the device ordinal's diminishing award;
7. authority is capped at the protocol maximum.

A participant must satisfy the authority threshold **and** the attestation gate to receive committee weight.

The newcomer ramp scales effective committee weight from 10% to 100% over the current fourteen-day period.

## Committee sampling

For each Monte Carlo committee draw, the simulator uses weighted sampling without replacement through an exponential-race construction.

For candidate weight `w` and uniform random value `U`:

```text
selection key = -ln(U) / w
```

The candidates with the smallest keys fill the committee seats. This produces weighted sampling without replacement while keeping the simulation independent of the production committee implementation.

The simulator reports two attack thresholds:

```text
blocking threshold
= minimum adversarial seats capable of preventing a 2/3 quorum

finality threshold
= adversarial seats meeting the 2/3 quorum itself
```

With the current 64-seat committee these are derived from the active quorum parameters rather than hard-coded constants.

## Report metrics

### Population

- total accounts/devices;
- adversarial accounts/devices;
- attested accounts;
- eligible accounts;
- eligible adversarial accounts;
- total and adversarial committee weight;
- adversarial weight share;
- mean honest/adversarial authority;
- Gini coefficient of committee weight;
- Herfindahl-Hirschman concentration index.

### Committee capture

- blocking-capture probability per committee;
- 95% Wilson interval for blocking capture;
- finality-control probability per committee;
- 95% Wilson interval for finality control;
- probability of at least one capture in a 144-epoch day under an independence approximation;
- mean adversarial seats;
- longest observed consecutive blocking/finality capture streak.

The per-day conversion is an approximation. Real epoch results may be correlated by shared membership, outages, or attack coordination.

### Liveness

- mean online committee seats;
- probability that at least quorum seats are online.

### Attack economics

Configured adversarial cohorts can include illustrative assumptions for:

- device purchase cost;
- account/bootstrap cost;
- monthly operating cost.

The report sums those values over the simulated duration and records the median number of days required for successful adversarial accounts to first cross the authority threshold.

These dollar values are **scenario assumptions**, not market-price claims and not a security guarantee.

## CI gate

CI runs:

```bash
go test ./...
go run ./cmd/pop-sim -scenario smoke -seed 1 -format json
go build ./cmd/pop-sim
```

The smoke scenario is intentionally small. Large attack sweeps belong in explicit research runs rather than every pull request.

## What v0.8 still does not prove

The model deliberately simplifies the future network. It does not yet simulate:

- peer-to-peer networking latency;
- block propagation;
- voting rounds or real finality messages;
- censorship strategies;
- adaptive corruption after committee revelation;
- bribery markets;
- geographic/provider correlation beyond the configured provider outage variables;
- real hardware resale value or operational market prices;
- legal/economic behavior around CARROT;
- a unique-human oracle, because v0.7 does not provide one.

The output should therefore be used to reject weak parameter choices and prioritize attacks, not to claim mathematical security.

## Exit criteria for v0.8

v0.8 is useful when we can answer, reproducibly:

- how quickly honest and adversarial accounts reach eligibility;
- how unattested bots compare with real-device farms;
- what attacker population produces material blocking/finality risk;
- how expensive that configured attack population is under stated assumptions;
- how authority decay and newcomer weighting change capture probability;
- how multi-device households differ from one-device farms;
- how provider outages and device churn affect liveness;
- whether the current 64-seat / 2/3 quorum assumptions remain reasonable enough to carry into the permissioned-testnet milestone.

Any parameter promoted to v0.9 should cite simulator results and the exact scenario/seed configuration used to justify it.
