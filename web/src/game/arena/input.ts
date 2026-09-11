import type { Direction, PlayerInput } from "./types.ts";

export type ControllerSnapshot = {
  axes: readonly number[];
  buttons: readonly boolean[];
};

export type ControllerActions = PlayerInput & {
  bombDown: boolean;
  remoteDown: boolean;
};

const DEAD_ZONE = 0.4;

export function controllerInput(snapshot?: ControllerSnapshot): ControllerActions {
  if (!snapshot) return { bombDown: false, remoteDown: false };

  const horizontal = snapshot.axes[0] ?? 0;
  const vertical = snapshot.axes[1] ?? 0;
  const dpadUp = Boolean(snapshot.buttons[12]);
  const dpadDown = Boolean(snapshot.buttons[13]);
  const dpadLeft = Boolean(snapshot.buttons[14]);
  const dpadRight = Boolean(snapshot.buttons[15]);

  let move: Direction | undefined;
  if (dpadUp) move = "up";
  else if (dpadDown) move = "down";
  else if (dpadLeft) move = "left";
  else if (dpadRight) move = "right";
  else if (Math.abs(horizontal) >= Math.abs(vertical) && Math.abs(horizontal) >= DEAD_ZONE) move = horizontal < 0 ? "left" : "right";
  else if (Math.abs(vertical) >= DEAD_ZONE) move = vertical < 0 ? "up" : "down";

  const bombDown = Boolean(snapshot.buttons[0]);
  const remoteDown = Boolean(snapshot.buttons[1]);
  return { move, bomb: bombDown, remote: remoteDown, bombDown, remoteDown };
}
