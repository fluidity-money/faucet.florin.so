import { NextRequest, NextResponse } from "next/server";
import { copySetCookie, forwardedHeaders, graphBaseUrl, responseJson } from "@/lib/backend";

async function proxy(request: NextRequest, method: "GET" | "DELETE") {
  try {
    const upstream = await fetch(`${graphBaseUrl()}/auth/x/session`, {
      method,
      headers: forwardedHeaders(request),
      cache: "no-store",
    });
    const headers = new Headers();
    copySetCookie(upstream, headers);
    return NextResponse.json(await responseJson(upstream), { status: upstream.status, headers });
  } catch {
    return NextResponse.json({ authenticated: false }, { status: 200 });
  }
}

export function GET(request: NextRequest) { return proxy(request, "GET"); }
export function DELETE(request: NextRequest) { return proxy(request, "DELETE"); }
