"use client";

import { FormEvent, useCallback, useMemo, useState } from "react";
import { readFaucetResult } from "@/lib/graphql";
import { isEthereumAddress } from "@/lib/validation";
import { Turnstile } from "./turnstile";

type EthereumProvider = { request: (args: { method: string; params?: unknown[] }) => Promise<unknown> };

declare global { interface Window { ethereum?: EthereumProvider } }

export function ClaimSuccess() {
  return (
    <div className="form-message success claim-success" role="status">
      <span>Funds sent. Head back to Florin to open your position.</span>
      <a className="btn btn--primary claim-success__cta" href="https://florin.so">
        Go back to Florin<span aria-hidden="true">→</span>
      </a>
    </div>
  );
}

export function Faucet() {
  const [address, setAddress] = useState("");
  const [turnstileToken, setTurnstileToken] = useState("");
  const [resetKey, setResetKey] = useState(0);
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState<{ kind: "error" | "success"; text: string } | null>(null);

  const connectWallet = async () => {
    setMessage(null);
    if (!window.ethereum) {
      setMessage({ kind: "error", text: "No browser wallet found. Paste an address instead." });
      return;
    }
    try {
      const accounts = await window.ethereum.request({ method: "eth_requestAccounts" });
      if (Array.isArray(accounts) && typeof accounts[0] === "string") setAddress(accounts[0]);
    } catch {
      setMessage({ kind: "error", text: "Wallet connection was cancelled." });
    }
  };

  const onToken = useCallback((token: string) => setTurnstileToken(token), []);
  const ready = useMemo(
    () => Boolean(isEthereumAddress(address) && turnstileToken && !pending),
    [address, pending, turnstileToken],
  );

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!ready) return;
    setPending(true);
    setMessage(null);
    try {
      const response = await fetch("/api/faucet", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ recipientAddress: address.trim(), turnstileToken }),
      });
      const result = await response.json().catch(() => ({}));
      const parsed = readFaucetResult(result);
      if (!response.ok) throw new Error(result.error || (!parsed.ok && parsed.message) || "The request failed.");
      if (!parsed.ok) throw new Error(parsed.message);
      setMessage({ kind: "success", text: "Request accepted. Your testnet funds are on the way." });
    } catch (reason) {
      setMessage({ kind: "error", text: reason instanceof Error ? reason.message : "The faucet request failed." });
    } finally {
      setPending(false);
      setTurnstileToken("");
      setResetKey((value) => value + 1);
    }
  };

  return (
    <form className="faucet-panel" onSubmit={submit}>
      <div className="faucet__rule">
        <span className="rule-pill"><span className="rule-pill__clock">◷</span> Once per 24h</span>
        <span className="rule-text">Free testnet funds for your address.</span>
      </div>

      <div className="fields">
        <div className="field">
          <span className="field__label" id="addr-label">Send to</span>
          <div className="field__control">
            <div className="addr">
              <input
                id="address"
                name="address"
                value={address}
                onChange={(event) => setAddress(event.target.value)}
                placeholder="0x…"
                autoComplete="off"
                spellCheck={false}
                aria-invalid={Boolean(address && !isEthereumAddress(address))}
                aria-labelledby="addr-label"
              />
              {address && isEthereumAddress(address) && <span className="addr__ok" aria-hidden="true">✓</span>}
              <button type="button" className="addr__btn" onClick={connectWallet}>Connect wallet</button>
            </div>
          </div>
          {address && !isEthereumAddress(address) && <small className="field-error">Enter a 42-character EVM address.</small>}
        </div>

        <div className="field">
          <span className="field__label" id="human-label">Human check</span>
          <div className="field__control">
            <Turnstile onToken={onToken} resetKey={resetKey} />
          </div>
        </div>
      </div>

      {message?.kind === "success" ? (
        <ClaimSuccess />
      ) : message ? (
        <div className="form-message error" role="status">{message.text}</div>
      ) : null}

      <div className="faucet__actions">
        <span className="meta">One claim per address every 2 hours</span>
        <button className="btn btn--primary btn--send" type="submit" disabled={!ready}>
          {pending ? "Sending…" : "Send tokens"}<span aria-hidden="true">→</span>
        </button>
      </div>
    </form>
  );
}
