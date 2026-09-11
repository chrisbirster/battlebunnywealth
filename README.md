# Battle Bunny Wealth

**Build the warren. Stack the carrots. Dominate the arena.**

Battle Bunny Wealth is an original mobile-first game concept combining an idle business/settlement loop with fast grid-based bomb battles. Underneath the game is an experimental blockchain research layer called **Proof of Play**: a protocol exploring whether long-lived, hardware-attested participation from ordinary devices can contribute to decentralized network security.

> Proof of Play is research code. This repository does not currently issue a cryptocurrency, promise financial returns, or implement production consensus.

## Architecture

```text
browser / future native shell
        |
        v
SolidJS 2 SPA + Solid Router + StyleX
        |
        | /api/v1
        v
Go net/http binary
        |
        +-- game/application APIs
        +-- Proof of Play protocol
        |      +-- epochs + challenges
        |      +-- participants + devices
        |      +-- attestation adapters
        |      +-- participation proofs
        |      +-- committee/finality (planned)
        |      +-- hash chain
        |
        +-- embedded Vite dist via go:embed
```

The production build is one Go binary containing the Vite-generated SPA.

## Development

```bash
cd web
npm install
npm run dev
```

In another terminal:

```bash
go run ./cmd/battlebunnywealth
```

Vite runs on `:5173` and proxies `/api` to the Go server on `:8080`.

Production build:

```bash
cd web && npm install && npm run build && cd ..
go build -o battlebunnywealth ./cmd/battlebunnywealth
./battlebunnywealth
```

Then open `http://localhost:8080`.

## Branch model

```text
feature/* -> dev -> release PR -> main -> vX.Y.Z tag
```

All normal implementation happens on feature branches. `dev` is integration. `main` contains released code only. Release tags must point to commits already merged to `main`; the release workflow enforces that rule.

## Documentation

Start with [`docs/README.md`](docs/README.md). The design docs cover the game loop, technical architecture, Proof of Play, identity/attestation, economy separation, security model, roadmap, and release process.
