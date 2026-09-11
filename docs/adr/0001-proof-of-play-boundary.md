# ADR 0001: Proof of Play is a separate protocol boundary

Status: Accepted

## Decision

Implement Proof of Play in `internal/proofofplay` without importing game/UI packages. The game may consume protocol events but protocol validity cannot depend on presentation concepts such as buildings, character XP, or bomb inventory.

## Why

This lets the team change the game economy without creating consensus forks and lets protocol simulations/tests run without the frontend.

## Consequence

Game rewards produced from participation are application projections of protocol events, not canonical consensus objects by default.
