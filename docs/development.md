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

Identity state is persisted to `data/identity.json` by default. Per-account game state lives in `data/players/<account-id>.json`.

The first authenticated account can automatically claim the old single-player `data/game-state.json` file. Override paths with:

```bash
BBWEALTH_IDENTITY_STATE=/tmp/bbw-identity.json \
BBWEALTH_GAME_DIR=/tmp/bbw-players \
BBWEALTH_GAME_STATE=/tmp/legacy-game-state.json \
go run ./cmd/battlebunnywealth
```

Production deployments must set the relying-party and exact allowed origin values for the deployed hostname, for example:

```bash
BBWEALTH_RP_ID=battlebunnywealth.com
BBWEALTH_ORIGINS=https://battlebunnywealth.com
```

`data/` is intentionally gitignored.

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
- Account-scoped game APIs require an authenticated passkey session.
- Browser session cookies are HttpOnly and SameSite=Strict.
- Treat ATProto `resolved-unverified` links as display metadata only; they grant no security authority.
- Treat v0.5 browser device keys as `unattested`; they grant no Proof-of-Play authority.
- Keep transport DTOs JSON-friendly but protocol domain types transport-agnostic when practical.
- Never expose private device keys.
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
