import assert from "node:assert/strict";
import test from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import HomePage from "../app/page";
import { ClaimSuccess, Faucet } from "../components/faucet";

test("successful claims offer a button back to Florin", () => {
  const html = renderToStaticMarkup(createElement(ClaimSuccess));

  assert.match(html, /Funds sent\. Head back to Florin to open your position\./);
  assert.match(html, /href="https:\/\/florin\.so"/);
  assert.match(html, />Go back to Florin<span aria-hidden="true">→<\/span><\/a>/);
});

test("the faucet can be used without connecting X", () => {
  const html = renderToStaticMarkup(createElement(Faucet));

  assert.doesNotMatch(html, /X account|Connect X|x\.com/i);
  assert.match(html, /<input[^>]+name="address"/);
  assert.doesNotMatch(html, /<input[^>]+name="address"[^>]+disabled/);
  assert.doesNotMatch(html, /Human check|captcha|challenges\.cloudflare\.com/i);
});

// The page is an async server component that reads the faucet's ETH balance.
// Render it offline: with the RPC unreachable it shows no low-ETH banner.
async function renderHomePage(): Promise<string> {
  const realFetch = globalThis.fetch;
  globalThis.fetch = (() => Promise.reject(new Error("offline"))) as typeof fetch;
  try {
    return renderToStaticMarkup(await HomePage());
  } finally {
    globalThis.fetch = realFetch;
  }
}

test("the page does not ask users to follow Florin on X", async () => {
  const html = await renderHomePage();

  assert.doesNotMatch(html, /follow[^<]*on X|x\.com/i);
});

test("the page points users to the Robinhood Chain faucet for gas", async () => {
  const html = await renderHomePage();

  assert.match(html, /href="https:\/\/faucet\.testnet\.chain\.robinhood\.com\/"/);
  assert.match(html, /Need ETH for gas\?/);
});
