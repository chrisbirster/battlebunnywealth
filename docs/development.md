# Development

## Toolchain

- Go 1.27.x
- Node.js 22
- Vite 8
- SolidJS 2 release candidate
- Solid Router v2 prerelease
- StyleX
- TypeScript

## Local loop

Terminal 1:

```bash
cd web
npm install
npm run dev
```

Terminal 2:

```bash
go run ./cmd/battlebunnywealth
```

The SPA runs at `http://localhost:5173`; Vite proxies `/api` to `http://127.0.0.1:8080`.

## Account + WebAuthn development

Local defaults are configured for passkeys on localhost:

```text
BBWEALTH_RP_ID=localhost
BBWEALTH_ORIGINS=http://localhost:8080,http://localhost:5173
```

Identity state is persisted to `data/identity.json` by default. Per-account game state lives in `data/players/<account-id>.json`. Proof-of-Play missions/authority persist to `data/proof-of-play.json`; v0.7 attestation challenges and provider records persist to `data/attestation.json`.

The first authenticated account can automatically claim the old single-player `data/game-state.json` file. Override paths with:

```bash
BBWEALTH_IDENTITY_STATE=/tmp/bbw-identity.json \
BBWEALTH_GAME_DIR=/tmp/bbw-players \
BBWEALTH_GAME_STATE=/tmp/legacy-game-state.json \
BBWEALTH_POP_STATE=/tmp/bbw-proof-of-play.json \
BBWEALTH_ATTESTATION_STATE=/tmp/bbw-attestation.json \
go run ./cmd/battlebunnywealth
```

Production deployments must set the relying-party and exact allowed origin values for the deployed hostname, for example:

```bash
BBWEALTH_RP_ID=battlebunnywealth.com
BBWEALTH_ORIGINS=https://battlebunnywealth.com
```

`data/` is intentionally gitignored.

## v0.6 mission testing

1. Create/sign into an account at `/account`.
2. Enroll the current browser as a device. The private P-256 key remains in IndexedDB.
3. Open `/proof-of-play`.
4. Request optional network duty.
5. Complete the mission before the five-minute expiry; the browser signs the canonical mission payload with the enrolled device key.
6. Repeat up to the four-mission daily ceiling and inspect the authority/service record.

Browser devices remain `unattested`. They can exercise prototype authority but do not satisfy v0.7's provider-backed device gate.

## v0.7 attestation configuration

Provider-backed attestation is performed by native clients. See `native/README.md`.

Apple configuration:

```text
BBWEALTH_APPLE_BUNDLE_ID=com.example.battlebunnywealth
BBWEALTH_APPLE_TEAM_ID=<team-id>
BBWEALTH_APPLE_ATTEST_ENV=development
```

The repository currently defines the Apple verifier boundary but does not ship a concrete App Attest server validator. The executable therefore fails Apple completion closed until one is supplied.

Android configuration:

```text
BBWEALTH_ANDROID_PACKAGE=com.example.battlebunnywealth
BBWEALTH_ANDROID_CERTIFICATES=<sha256-cert-digest>[,<sha256-cert-digest>...]
BBWEALTH_ANDROID_REQUIRE_STRONG_INTEGRITY=1
BBWEALTH_PLAY_INTEGRITY_ACCESS_TOKEN=<short-lived-oauth-access-token>
```

`PlayIntegrityHTTPDecoder` calls Google `decodeIntegrityToken`. The static access-token environment variable is suitable only for development/integration work; a deployment needs a renewable service-account/ADC token source.

For deterministic local adapter tests only:

```text
BBWEALTH_DEV_CONTROLS=1
```

This enables the `development` attestation provider. That provider is permanently non-hardware-backed and non-eligible for the attested committee gate.

## Production parity

```bash
cd web
npm run test:arena
npm run typecheck
npm run build
cd ..
go test ./...
go vet ./...
go build ./cmd/battlebunnywealth
```

The Go binary embeds whatever is currently in `web/dist`.

## API rules

- Prefix HTTP APIs with `/api/v1`.
- Treat game economy mutations as server-authoritative; browser-calculated balances are never trusted.
- Account-scoped game and personal Proof-of-Play APIs require an authenticated passkey session.
- Browser session cookies are HttpOnly and SameSite=Strict.
- Treat ATProto `resolved-unverified` links as display metadata only; they grant no security authority.
- Browser device keys remain unattested unless a native provider flow verifies the corresponding enrolled device record.
- Never treat an attested device as proof of one unique person.
- Keep platform attestation behind provider interfaces; Apple/Google are trust inputs, not protocol identity.
- Keep transport DTOs JSON-friendly but protocol domain types transport-agnostic when practical.
- Never expose private device keys.
- Never publish raw account IDs/device public keys from committee prototypes.
- Version signed/protocol structures independently of REST API versions.
- Add tests for every consensus validation rule.

## Branches

Create work from `dev`:

```bash
git switch dev
git pull
git switch -c feature/<short-name>
```

Open PRs from `feature/*` into `dev`. Do not open ordinary feature PRs directly against `main`.
