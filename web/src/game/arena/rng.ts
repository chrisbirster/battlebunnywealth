export function nextRandom(state: number): { state: number; value: number } {
  let x = state | 0;
  x ^= x << 13;
  x ^= x >>> 17;
  x ^= x << 5;
  const next = x >>> 0 || 0x9e3779b9;
  return { state: next, value: next / 0x1_0000_0000 };
}

export function randomInt(state: number, maxExclusive: number): { state: number; value: number } {
  const next = nextRandom(state);
  return { state: next.state, value: Math.floor(next.value * maxExclusive) };
}
