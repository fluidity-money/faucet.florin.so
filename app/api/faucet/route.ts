import { NextRequest, NextResponse } from "next/server";
import { forwardedHeaders, graphBaseUrl, responseJson } from "@/lib/backend";
import { buildFaucetRequest } from "@/lib/graphql";
import { isEthereumAddress } from "@/lib/validation";

export async function POST(request: NextRequest) {
  let input: { recipientAddress?: string; turnstileToken?: string };
  try { input = await request.json(); }
  catch { return NextResponse.json({ error: "Invalid request body." }, { status: 400 }); }

  const recipientAddress = input.recipientAddress?.trim() ?? "";
  const turnstileToken = input.turnstileToken ?? "";
  if (!isEthereumAddress(recipientAddress)) return NextResponse.json({ error: "Enter a valid EVM address." }, { status: 400 });
  if (!turnstileToken) return NextResponse.json({ error: "Complete the anti-bot check." }, { status: 400 });

  try {
    const upstream = await fetch(`${graphBaseUrl()}/query`, {
      method: "POST",
      headers: forwardedHeaders(request, true),
      body: JSON.stringify(buildFaucetRequest({ recipientAddress, turnstileToken })),
      cache: "no-store",
    });
    return NextResponse.json(await responseJson(upstream), { status: upstream.status });
  } catch {
    return NextResponse.json({ error: "The faucet service is unavailable." }, { status: 502 });
  }
}
