import assert from "node:assert/strict";
import test from "node:test";
import { isEthereumAddress, normalizeXHandle, validateXHandle } from "../lib/validation";

test("normalizes a pasted X handle", () => {
  assert.equal(normalizeXHandle("  @Florin_Faucet  "), "Florin_Faucet");
});

test("accepts only X-compatible handles", () => {
  assert.equal(validateXHandle("florin_123"), true);
  assert.equal(validateXHandle(""), false);
  assert.equal(validateXHandle("sixteen_chars_now"), false);
  assert.equal(validateXHandle("not-valid"), false);
});

test("accepts an EVM address and rejects malformed values", () => {
  assert.equal(isEthereumAddress("0x000000000000000000000000000000000000dEaD"), true);
  assert.equal(isEthereumAddress("0x1234"), false);
  assert.equal(isEthereumAddress("0000000000000000000000000000000000000000"), false);
});
