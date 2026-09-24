import { NextRequest, NextResponse } from "next/server";
import { copySetCookie, forwardedHeaders, graphBaseUrl, responseJson } from "@/lib/backend";
import { safeReturnTo } from "@/lib/x-auth";

export async function GET(request: NextRequest) {
  const returnTo = safeReturnTo(request.nextUrl.searchParams.get("returnTo"));
  const redirectUri = `${request.nextUrl.origin}/auth/x/callback`;
  const url = new URL(`${graphBaseUrl()}/auth/x/start`);
  url.searchParams.set("redirectUri", redirectUri);
  url.searchParams.set("returnTo", returnTo);

  try {
    const upstream = await fetch(url, {
      headers: forwardedHeaders(request),
      cache: "no-store",
      redirect: "manual",
    });
    const headers = new Headers();
    copySetCookie(upstream, headers);
    const location = upstream.headers.get("location");
    if (location) return NextResponse.json({ authorizationUrl: location }, { status: 200, headers });
    return NextResponse.json(await responseJson(upstream), { status: upstream.status, headers });
  } catch {
    return NextResponse.json({ error: "X verification is unavailable." }, { status: 502 });
  }
}
