import assert from "node:assert/strict";
import test from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { ClaimSuccess } from "../components/faucet";

test("successful claims offer a button back to Florin", () => {
  const html = renderToStaticMarkup(createElement(ClaimSuccess));

  assert.match(html, /Funds sent\. Head back to Florin to open your position\./);
  assert.match(html, /href="https:\/\/florin\.so"/);
  assert.match(html, />Go back to Florin<span aria-hidden="true">→<\/span><\/a>/);
});
