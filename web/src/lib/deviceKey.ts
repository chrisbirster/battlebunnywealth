export async function createBrowserDeviceKey(): Promise<string> {
  const pair = await crypto.subtle.generateKey({ name: "ECDSA", namedCurve: "P-256" }, false, ["sign", "verify"]);
  const spki = await crypto.subtle.exportKey("spki", pair.publicKey);
  const encoded = toB64(spki);
  await persistDeviceKey(pair.privateKey, encoded);
  return encoded;
}

export async function currentBrowserDevicePublicKey(): Promise<string | undefined> {
  const db = await openDB();
  try {
    return await new Promise<string | undefined>((resolve, reject) => {
      const tx = db.transaction("device-keys", "readonly");
      const req = tx.objectStore("device-keys").get("current-spki");
      req.onsuccess = () => resolve(typeof req.result === "string" ? req.result : undefined);
      req.onerror = () => reject(req.error);
    });
  } finally {
    db.close();
  }
}

export async function signCurrentDevicePayload(payload: string): Promise<string> {
  const db = await openDB();
  try {
    const key = await new Promise<CryptoKey | undefined>((resolve, reject) => {
      const tx = db.transaction("device-keys", "readonly");
      const req = tx.objectStore("device-keys").get("current");
      req.onsuccess = () => resolve(req.result instanceof CryptoKey ? req.result : undefined);
      req.onerror = () => reject(req.error);
    });
    if (!key) throw new Error("This browser does not have an enrolled device key. Enroll it from Account first.");
    const signature = await crypto.subtle.sign({ name: "ECDSA", hash: "SHA-256" }, key, new TextEncoder().encode(payload));
    return toB64(signature);
  } finally {
    db.close();
  }
}

async function persistDeviceKey(key: CryptoKey, publicKeySpki: string): Promise<void> {
  const db = await openDB();
  await new Promise<void>((resolve, reject) => {
    const tx = db.transaction("device-keys", "readwrite");
    const store = tx.objectStore("device-keys");
    store.put(key, "current");
    store.put(publicKeySpki, "current-spki");
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error);
  });
  db.close();
}

function openDB(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open("battle-bunny-wealth", 1);
    req.onupgradeneeded = () => { if (!req.result.objectStoreNames.contains("device-keys")) req.result.createObjectStore("device-keys"); };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error);
  });
}

function toB64(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (const byte of bytes) binary += String.fromCharCode(byte);
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
}
