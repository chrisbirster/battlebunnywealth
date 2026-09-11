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

The v0.3 single-player game state persists to `data/game-state.json`. Override the location with:

```bash
BBWEALTH_GAME_STATE=/tmp/battle-bunny-state.json go run ./cmd/battlebunnywealth
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
