import type { NextRequest } from "next/server";

const DEFAULT_GRAPH_URL = "https://faucet-graph.florin.so";

export function graphBaseUrl(): string {
  return (process.env.FAUCET_GRAPH_URL || DEFAULT_GRAPH_URL).replace(/\/$/, "");
}

export function forwardedHeaders(request: NextRequest, includeJson = false): Headers {
  const headers = new Headers();
  const cookie = request.headers.get("cookie");
  if (cookie) headers.set("cookie", cookie);
  const forwardedFor = request.headers.get("x-forwarded-for");
  if (forwardedFor) headers.set("x-forwarded-for", forwardedFor);
  headers.set("x-forwarded-proto", request.nextUrl.protocol.replace(":", ""));
  headers.set("x-forwarded-host", request.nextUrl.host);
  const userAgent = request.headers.get("user-agent");
  if (userAgent) headers.set("user-agent", userAgent);
  if (includeJson) headers.set("content-type", "application/json");
  return headers;
}

export function copySetCookie(source: Response, destination: Headers): void {
  const headers = source.headers as Headers & { getSetCookie?: () => string[] };
  const values = headers.getSetCookie?.() ?? [];
  if (values.length) {
    for (const value of values) destination.append("set-cookie", value);
    return;
  }
  const single = source.headers.get("set-cookie");
  if (single) destination.append("set-cookie", single);
}

export async function responseJson(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) return {};
  try { return JSON.parse(text); } catch { return { error: text }; }
}
