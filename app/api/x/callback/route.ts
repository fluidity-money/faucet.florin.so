import { NextRequest, NextResponse } from "next/server";
import { copySetCookie, forwardedHeaders, graphBaseUrl, responseJson } from "@/lib/backend";

export async function POST(request: NextRequest) {
  let body: unknown;
  try { body = await request.json(); }
  catch { return NextResponse.json({ error: "Invalid callback." }, { status: 400 }); }

  try {
    const upstream = await fetch(`${graphBaseUrl()}/auth/x/callback`, {
      method: "POST",
      headers: forwardedHeaders(request, true),
      body: JSON.stringify(body),
      cache: "no-store",
    });
    const headers = new Headers();
    copySetCookie(upstream, headers);
    return NextResponse.json(await responseJson(upstream), { status: upstream.status, headers });
  } catch {
    return NextResponse.json({ error: "Could not finish X verification." }, { status: 502 });
  }
}
