export type NetworkStatus = {
  protocol: string;
  phase: string;
  epoch: number;
  chainHeight: number;
  committeeSize: number;
  quorum: string;
  epochSeconds: number;
  challengeTtlSeconds: number;
  maxHumanChallengesPerDay: number;
  attestationProviders: string[];
  authorityMode: string;
  productionCommitteeEnabled: boolean;
};

export type NetworkMission = {
  id: string;
  deviceId: string;
  deviceAttestation: string;
  kind: string;
  npc: string;
  title: string;
  briefing: string;
  operation: string;
  epoch: number;
  challenge: string;
  checkpointHash: string;
  signedPayload: string;
  authorityAward: number;
  status: string;
  issuedAt: string;
  expiresAt: string;
  completedAt?: string;
};

export type MissionCompletion = {
  missionId: string;
  deviceId: string;
  epoch: number;
  kind: string;
  authorityAward: number;
  signatureDigest: string;
  completedAt: string;
};

export type AuthoritySnapshot = {
  authority: {
    score: number;
    maximum: number;
    missionsCompleted: number;
    eligibilityThreshold: number;
    prototypeEligible: boolean;
    productionEligible: boolean;
    committeeWeight: number;
    newcomerWeightPercent: number;
    decayGraceSeconds: number;
    decayPerDayBasisPoints: number;
    firstMissionAt?: string;
    lastMissionAt?: string;
  };
  currentMission?: NetworkMission;
  recentCompletions: MissionCompletion[];
  dailyMissionsUsed: number;
  dailyMissionLimit: number;
  attestationGate: string;
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) } });
  if (!response.ok) {
    const body = await response.json().catch(() => ({ error: `HTTP ${response.status}` })) as { error?: string };
    const error = new Error(body.error ?? `HTTP ${response.status}`) as Error & { status?: number };
    error.status = response.status;
    throw error;
  }
  return response.json() as Promise<T>;
}

export const proofApi = {
  status: () => request<NetworkStatus>("/api/v1/proof-of-play"),
  me: () => request<AuthoritySnapshot>("/api/v1/proof-of-play/me"),
  issueMission: (deviceId: string) => request<AuthoritySnapshot>("/api/v1/proof-of-play/missions", { method: "POST", body: JSON.stringify({ deviceId }) }),
  completeMission: (missionId: string, deviceId: string, signature: string) => request<AuthoritySnapshot>(`/api/v1/proof-of-play/missions/${encodeURIComponent(missionId)}/complete`, { method: "POST", body: JSON.stringify({ deviceId, signature }) }),
};
