import assert from "node:assert/strict";
import test from "node:test";

import { controllerInput } from "./input.ts";

test("controller d-pad maps to cardinal movement", () => {
  const buttons = Array<boolean>(16).fill(false);
  buttons[12] = true;
  assert.equal(controllerInput({ axes: [0, 0], buttons }).move, "up");

  buttons.fill(false);
  buttons[15] = true;
  assert.equal(controllerInput({ axes: [0, 0], buttons }).move, "right");
});

test("controller sticks respect the dead zone and dominant axis", () => {
  assert.equal(controllerInput({ axes: [0.2, -0.3], buttons: [] }).move, undefined);
  assert.equal(controllerInput({ axes: [0.8, -0.6], buttons: [] }).move, "right");
  assert.equal(controllerInput({ axes: [0.1, 0.9], buttons: [] }).move, "down");
});

test("standard face buttons map to bomb and remote", () => {
  const buttons = [true, true];
  const input = controllerInput({ axes: [0, 0], buttons });
  assert.equal(input.bombDown, true);
  assert.equal(input.remoteDown, true);
});
