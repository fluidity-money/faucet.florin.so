export type XCallbackPayload = {
  code: string;
  state: string;
  redirectUri: string;
};

export function callbackPayload(
  params: URLSearchParams,
  redirectUri: string,
): XCallbackPayload {
  const code = params.get("code");
  const state = params.get("state");
  if (!code || !state) throw new Error("X did not return a complete authorization response.");
  return { code, state, redirectUri };
}

export function safeReturnTo(value: string | null | undefined): string {
  if (!value || !value.startsWith("/") || value.startsWith("//")) return "/";
  try {
    const decoded = decodeURIComponent(value);
    if (decoded.startsWith("//") || /[\\\\\u0000-\u001f\u007f]/.test(decoded)) return "/";
    const base = new URL("https://faucet.invalid");
    if (new URL(value, base).origin !== base.origin) return "/";
  } catch {
    return "/";
  }
  return value;
}
