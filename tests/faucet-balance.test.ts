import assert from "node:assert/strict";
import test from "node:test";
import { encodeBalancesCall, sumBalances } from "../lib/faucet-balance";

const word = (n: number | bigint) => n.toString(16).padStart(64, "0");

test("encodes one getEthBalance call per wallet for aggregate3", () => {
  const data = encodeBalancesCall(["0x000000000000000000000000000000000000dEaD"]);
  assert.equal(data.slice(0, 10), "0x82ad56cb");
  assert.equal(data.length, 2 + 8 + 64 * 3 + 64 * 6);
  assert.ok(data.includes("4d2301cc000000000000000000000000000000000000000000000000000000000000dead"));
});

test("sums successful balances and skips failed calls", () => {
  const tuple = (ok: boolean, value: bigint) => word(ok ? 1 : 0) + word(64) + word(32) + word(value);
  const result = "0x" + word(32) + word(3) + word(96) + word(96 + 128) + word(96 + 256)
    + tuple(true, 10n ** 16n) + tuple(false, 99n) + tuple(true, 2n * 10n ** 16n);
  assert.equal(sumBalances(result), 3n * 10n ** 16n);
});
