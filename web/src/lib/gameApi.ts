export type PlayerProfile = {
  name: string;
  callsign: string;
  fur: string;
  ears: string;
  uniform: string;
  cosmetics: string[];
};

export type Business = {
  id: string;
  name: string;
  description: string;
  icon: string;
  level: number;
  incomePerSecond: number;
  upgradeCost: number;
  synergy: string;
};

export type GameSnapshot = {
  schemaVersion: number;
  player: PlayerProfile;
  wallet: { bunnyBucks: number };
  businesses: Business[];
  season: { id: string; earnings: number; turnIns: number };
  incomePerSecond: number;
  accruedBunnyBucks: number;
  offlineCapSeconds: number;
  seasonTurnInTarget: number;
  rankedPowerAffected: boolean;
  onboardingStep: number;
};

export type Standing = { rank: number; name: string; callsign: string; bunnyBucks: number };

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
  });
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: `HTTP ${response.status}` })) as { error?: string };
    throw new Error(body.error ?? `HTTP ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export const gameApi = {
  state: () => request<GameSnapshot>("/api/v1/game/state"),
  updateProfile: (profile: PlayerProfile) => request<GameSnapshot>("/api/v1/game/profile", { method: "PUT", body: JSON.stringify(profile) }),
  upgrade: (id: string) => request<GameSnapshot>(`/api/v1/game/businesses/${encodeURIComponent(id)}/upgrade`, { method: "POST" }),
  advanceOnboarding: () => request<GameSnapshot>("/api/v1/game/onboarding/advance", { method: "POST" }),
  turnIn: () => request<GameSnapshot>("/api/v1/game/season/turn-in", { method: "POST" }),
  standings: () => request<Standing[]>("/api/v1/game/standings"),
};
