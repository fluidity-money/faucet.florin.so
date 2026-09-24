import assert from "node:assert/strict";
import test from "node:test";
import { callbackPayload, safeReturnTo } from "../lib/x-auth";

test("builds a server exchange payload from an X callback", () => {
  const params = new URLSearchParams("code=abc123&state=state456");
  assert.deepEqual(callbackPayload(params, "https://faucet.florin.so/auth/x/callback"), {
    code: "abc123",
    state: "state456",
    redirectUri: "https://faucet.florin.so/auth/x/callback",
  });
});

test("rejects incomplete callback parameters", () => {
  assert.throws(() => callbackPayload(new URLSearchParams("code=abc123"), "https://example.test/callback"));
});

test("allows only local return paths", () => {
  assert.equal(safeReturnTo("/"), "/");
  assert.equal(safeReturnTo("/claim?from=x"), "/claim?from=x");
  assert.equal(safeReturnTo("https://evil.example"), "/");
  assert.equal(safeReturnTo("//evil.example"), "/");
});
