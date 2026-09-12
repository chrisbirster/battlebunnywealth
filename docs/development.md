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

Identity state is persisted to `data/identity.json` by default. Per-account game state lives in `data/players/<account-id>.json`. Proof-of-Play missions and authority persist separately to `data/proof-of-play.json`.

The first authenticated account can automatically claim the old single-player `data/game-state.json` file. Override paths with:

```bash
BBWEALTH_IDENTITY_STATE=/tmp/bbw-identity.json \
BBWEALTH_GAME_DIR=/tmp/bbw-players \
BBWEALTH_GAME_STATE=/tmp/legacy-game-state.json \
BBWEALTH_POP_STATE=/tmp/bbw-proof-of-play.json \
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

The current browser key is still `unattested`. v0.6 may use it to exercise **prototype** authority and committee weighting, but production committee eligibility is hard-disabled until v0.7 attestation.

If a browser device was enrolled before v0.6, its IndexedDB record does not contain the new cached public-key identifier used to match the local key to the server device. Re-enroll that browser once for mission testing.

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
- Treat v0.5/v0.6 browser device keys as `unattested`; they may exercise prototype authority but never production committee eligibility.
- Keep transport DTOs JSON-friendly but protocol domain types transport-agnostic when practical.
- Never expose private device keys.
- Never publish raw account IDs/device public keys from the v0.6 committee prototype.
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
