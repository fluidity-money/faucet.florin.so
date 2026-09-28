import assert from "node:assert/strict";
import test from "node:test";
import { buildFaucetRequest, readFaucetResult } from "../lib/graphql";

test("builds the requestToken GraphQL operation expected by gqlgen", () => {
  const body = buildFaucetRequest({
    xHandle: "florin",
    recipientAddress: "0x000000000000000000000000000000000000dEaD",
    turnstileToken: "turnstile-proof",
  });

  assert.equal(body.operationName, "RequestToken");
  assert.match(body.query, /requestToken/);
  assert.deepEqual(body.variables, {
    xHandle: "florin",
    recipientAddress: "0x000000000000000000000000000000000000dEaD",
    turnstileToken: "turnstile-proof",
  });
});

test("requires an affirmative result and no GraphQL errors", () => {
  assert.deepEqual(readFaucetResult({ data: { requestToken: true } }), { ok: true });
  assert.deepEqual(
    readFaucetResult({ errors: [{ message: "daily limit reached" }] }),
    { ok: false, message: "daily limit reached", authenticationRequired: false },
  );
  assert.deepEqual(readFaucetResult({ data: { requestToken: false } }), {
    ok: false,
    message: "The faucet did not accept this request.",
    authenticationRequired: false,
  });
});

test("identifies a rejected X session so the UI can disconnect it", () => {
  assert.deepEqual(
    readFaucetResult({ errors: [{ message: "Connect and verify your X account first." }] }),
    {
      ok: false,
      message: "Connect and verify your X account first.",
      authenticationRequired: true,
    },
  );
});
