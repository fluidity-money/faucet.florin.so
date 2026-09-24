import { NextRequest, NextResponse } from "next/server";
import { forwardedHeaders, graphBaseUrl, responseJson } from "@/lib/backend";
import { buildFaucetRequest } from "@/lib/graphql";
import { isEthereumAddress, normalizeXHandle, validateXHandle } from "@/lib/validation";

export async function POST(request: NextRequest) {
  let input: { xHandle?: string; recipientAddress?: string; turnstileToken?: string };
  try { input = await request.json(); }
  catch { return NextResponse.json({ error: "Invalid request body." }, { status: 400 }); }

  const xHandle = normalizeXHandle(input.xHandle ?? "");
  const recipientAddress = input.recipientAddress?.trim() ?? "";
  const turnstileToken = input.turnstileToken ?? "";
  if (!validateXHandle(xHandle)) return NextResponse.json({ error: "Connect a valid X account first." }, { status: 400 });
  if (!isEthereumAddress(recipientAddress)) return NextResponse.json({ error: "Enter a valid EVM address." }, { status: 400 });
  if (!turnstileToken) return NextResponse.json({ error: "Complete the anti-bot check." }, { status: 400 });

  try {
    const upstream = await fetch(`${graphBaseUrl()}/query`, {
      method: "POST",
      headers: forwardedHeaders(request, true),
      body: JSON.stringify(buildFaucetRequest({ xHandle, recipientAddress, turnstileToken })),
      cache: "no-store",
    });
    return NextResponse.json(await responseJson(upstream), { status: upstream.status });
  } catch {
    return NextResponse.json({ error: "The faucet service is unavailable." }, { status: 502 });
  }
}
