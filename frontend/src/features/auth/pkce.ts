function dec2hex(dec: number): string {
  return dec.toString(16).padStart(2, '0');
}

export function generateCodeVerifier(): string {
  const array = new Uint32Array(56 / 2);
  window.crypto.getRandomValues(array);
  return Array.from(array, dec2hex).join('');
}

function sha256(plain: string): Promise<ArrayBuffer> {
  const encoder = new TextEncoder();
  const data = encoder.encode(plain);
  return window.crypto.subtle.digest('SHA-256', data);
}

function base64urlencode(a: ArrayBuffer): string {
  let str = "";
  const bytes = new Uint8Array(a);
  const len = bytes.byteLength;
  for (let i = 0; i < len; i++) {
    str += String.fromCharCode(bytes[i]);
  }
  return btoa(str)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

export async function generateCodeChallenge(verifier: string): Promise<string> {
  const hashed = await sha256(verifier);
  return base64urlencode(hashed);
}

export function generateState(): string {
  const array = new Uint32Array(28 / 2);
  window.crypto.getRandomValues(array);
  return Array.from(array, dec2hex).join('');
}
