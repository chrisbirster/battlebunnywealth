export type ATProtoBinding = { did: string; handle?: string; pds?: string; status: string; boundAt: string };
export type Device = { id: string; name: string; platform: string; publicKeySpki: string; status: string; attestationStatus: string; enrolledAt: string; revokedAt?: string };
export type Account = { id: string; displayName: string; createdAt: string; passkeyCount: number; atproto?: ATProtoBinding; devices: Device[] };

type RegistrationBegin = { ceremony: string; publicKey: any };
type LoginBegin = { ceremony: string; publicKey: any };

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

export const authApi = {
  me: () => request<Account>("/api/v1/auth/me"),
  logout: () => request<{ ok: boolean }>("/api/v1/auth/logout", { method: "POST", body: "{}" }),
  register: async (displayName: string) => {
    const begin = await request<RegistrationBegin>("/api/v1/auth/passkey/register/begin", { method: "POST", body: JSON.stringify({ displayName }) });
    const credential = await navigator.credentials.create({ publicKey: creationOptions(begin.publicKey) }) as PublicKeyCredential | null;
    if (!credential) throw new Error("Passkey creation was cancelled.");
    return request<Account>("/api/v1/auth/passkey/register/finish", { method: "POST", body: JSON.stringify({ ceremony: begin.ceremony, credential: serializeRegistration(credential) }) });
  },
  addPasskey: async () => {
    const begin = await request<RegistrationBegin>("/api/v1/auth/passkey/register/begin", { method: "POST", body: JSON.stringify({ displayName: "" }) });
    const credential = await navigator.credentials.create({ publicKey: creationOptions(begin.publicKey) }) as PublicKeyCredential | null;
    if (!credential) throw new Error("Passkey creation was cancelled.");
    return request<Account>("/api/v1/auth/passkey/register/finish", { method: "POST", body: JSON.stringify({ ceremony: begin.ceremony, credential: serializeRegistration(credential) }) });
  },
  login: async () => {
    const begin = await request<LoginBegin>("/api/v1/auth/passkey/login/begin", { method: "POST", body: "{}" });
    const credential = await navigator.credentials.get({ publicKey: loginOptions(begin.publicKey) }) as PublicKeyCredential | null;
    if (!credential) throw new Error("Passkey sign-in was cancelled.");
    return request<Account>("/api/v1/auth/passkey/login/finish", { method: "POST", body: JSON.stringify({ ceremony: begin.ceremony, credential: serializeAssertion(credential) }) });
  },
  linkATProto: (did: string) => request<Account>("/api/v1/account/atproto", { method: "PUT", body: JSON.stringify({ did }) }),
  unlinkATProto: () => request<Account>("/api/v1/account/atproto", { method: "DELETE", body: "{}" }),
  enrollDevice: (name: string, platform: string, publicKeySpki: string) => request<Account>("/api/v1/account/devices", { method: "POST", body: JSON.stringify({ name, platform, publicKeySpki }) }),
  revokeDevice: (id: string) => request<Account>(`/api/v1/account/devices/${encodeURIComponent(id)}/revoke`, { method: "POST", body: "{}" }),
};

function creationOptions(input: any): PublicKeyCredentialCreationOptions {
  return { ...input, challenge: fromB64(input.challenge), user: { ...input.user, id: fromB64(input.user.id) }, excludeCredentials: (input.excludeCredentials ?? []).map((item: any) => ({ ...item, id: fromB64(item.id) })) } as PublicKeyCredentialCreationOptions;
}
function loginOptions(input: any): PublicKeyCredentialRequestOptions { return { ...input, challenge: fromB64(input.challenge) } as PublicKeyCredentialRequestOptions; }
function serializeRegistration(credential: PublicKeyCredential) { const response = credential.response as AuthenticatorAttestationResponse; return { id: credential.id, rawId: toB64(credential.rawId), type: credential.type, response: { clientDataJSON: toB64(response.clientDataJSON), attestationObject: toB64(response.attestationObject), transports: response.getTransports?.() ?? [] } }; }
function serializeAssertion(credential: PublicKeyCredential) { const response = credential.response as AuthenticatorAssertionResponse; return { id: credential.id, rawId: toB64(credential.rawId), type: credential.type, response: { clientDataJSON: toB64(response.clientDataJSON), authenticatorData: toB64(response.authenticatorData), signature: toB64(response.signature), userHandle: response.userHandle ? toB64(response.userHandle) : "" } }; }
function toB64(buffer: ArrayBuffer): string { const bytes = new Uint8Array(buffer); let binary = ""; for (const byte of bytes) binary += String.fromCharCode(byte); return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, ""); }
function fromB64(value: string): Uint8Array { const padded = value.replace(/-/g, "+").replace(/_/g, "/") + "===".slice((value.length + 3) % 4); const binary = atob(padded); const out = new Uint8Array(binary.length); for (let i = 0; i < binary.length; i++) out[i] = binary.charCodeAt(i); return out; }
