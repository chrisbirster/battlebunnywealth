# Release process

Battle Bunny Wealth uses an integration branch and a release-only main branch.

```text
feature/foo ----\
feature/bar -----+--> dev --> PR: dev -> main --> merge --> vX.Y.Z tag
hotfix/* --------/                         ^
                                      released code
```

## Rules

1. Branch new work from `dev`.
2. PR feature branches back into `dev`.
3. CI must pass before merge.
4. When `dev` is release-ready, open one release PR from `dev` to `main`.
5. Merge the release PR.
6. Create an annotated semantic-version tag on the resulting `main` commit.
7. Push the tag. The Release workflow verifies that the tagged commit is on `main`, rebuilds/tests, and publishes binaries.

Example:

```bash
git switch dev
git pull
# after release PR is merged
git switch main
git pull
git tag -a v0.1.0 -m "Battle Bunny Wealth v0.1.0"
git push origin v0.1.0
```

## Version intent

- `v0.x`: game/protocol research; breaking changes expected.
- `v1.0`: requires a stable playable core. It does **not** imply Proof of Play is a production financial network unless explicitly documented.

## Hotfixes

Create `hotfix/<name>` from `main`, fix/test, PR to `main`, tag a patch release, then merge or cherry-pick the same fix back into `dev` immediately so branches do not diverge.
