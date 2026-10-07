import assert from "node:assert/strict";
import test from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import HomePage from "../app/page";
import { ClaimSuccess, Faucet } from "../components/faucet";
import { TURNSTILE_ACTION, TURNSTILE_SITE_KEY } from "../components/turnstile";

test("successful claims offer a button back to Florin", () => {
  const html = renderToStaticMarkup(createElement(ClaimSuccess));

  assert.match(html, /Funds sent\. Head back to Florin to open your position\./);
  assert.match(html, /href="https:\/\/florin\.so"/);
  assert.match(html, />Go back to Florin<span aria-hidden="true">→<\/span><\/a>/);
});

test("the faucet uses the existing Cloudflare Turnstile widget", () => {
  assert.equal(TURNSTILE_SITE_KEY, "0x4AAAAAAFCW3wdeeOUrUHOs");
  assert.equal(TURNSTILE_ACTION, "faucet");
});

test("the faucet can be used without connecting X", () => {
  const html = renderToStaticMarkup(createElement(Faucet));

  assert.doesNotMatch(html, /X account|Connect X|x\.com/i);
  assert.match(html, /<input[^>]+name="address"/);
  assert.doesNotMatch(html, /<input[^>]+name="address"[^>]+disabled/);
  assert.match(html, /Human check/);
});

test("the page does not ask users to follow Florin on X", () => {
  const html = renderToStaticMarkup(createElement(HomePage));

  assert.doesNotMatch(html, /follow[^<]*on X|x\.com/i);
});
