# Fly.io public-testnet deployment

Fly.io is the only deployment target for Battle Bunny Wealth.

The four checked-in configs intentionally represent four separate Fly apps in four regions: `iad`, `ord`, `dfw`, and `lax`. Pass the app name with `-a`; app names are not hard-coded because Fly app names are globally unique.

Before deployment each app needs:

1. a `pop_data` Fly Volume in the config's region;
2. `POP_NODE_CONFIG` secret containing base64-encoded `node.json`;
3. `POP_NODE_KEY` secret containing the base64-encoded node key file.

The configs mount those secrets at `/config/node.json` and `/config/node.key`, persist consensus state at `/data`, disable automatic stopping, expose port `9101` through Fly HTTPS, and health-check `/v1/public/status`.

Use `deploy/public-testnet/README.md` for the full bootstrap sequence.
