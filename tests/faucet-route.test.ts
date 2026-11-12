import assert from "node:assert/strict";
import test from "node:test";
import { NextRequest } from "next/server";
import { POST } from "../app/api/faucet/route";
import { buildFaucetRequest } from "../lib/graphql";

test("claims require only a valid recipient address", async (t) => {
  const recipientAddress = "0x000000000000000000000000000000000000dEaD";
  const upstream = t.mock.method(globalThis, "fetch", async (_url: unknown, options?: RequestInit) => {
    assert.deepEqual(JSON.parse(String(options?.body)), buildFaucetRequest({ recipientAddress }));
    return Response.json({ data: { requestToken: true } });
  });
  const response = await POST(new NextRequest("https://faucet.florin.so/api/faucet", {
    method: "POST",
    body: JSON.stringify({ recipientAddress }),
  }));
  assert.equal(response.status, 200);
  assert.deepEqual(await response.json(), { data: { requestToken: true } });
  assert.equal(upstream.mock.callCount(), 1);
});

test("invalid recipients are rejected before contacting the backend", async (t) => {
  const upstream = t.mock.method(globalThis, "fetch", async () => {
    throw new Error("unexpected backend request");
  });
  const response = await POST(new NextRequest("https://faucet.florin.so/api/faucet", {
    method: "POST",
    body: JSON.stringify({ recipientAddress: "invalid" }),
  }));
  assert.equal(response.status, 400);
  assert.equal(upstream.mock.callCount(), 0);
});
