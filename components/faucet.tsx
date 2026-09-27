"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { readFaucetResult } from "@/lib/graphql";
import { isEthereumAddress } from "@/lib/validation";
import { Turnstile } from "./turnstile";

type XSession = { authenticated: boolean; handle?: string };
type EthereumProvider = { request: (args: { method: string; params?: unknown[] }) => Promise<unknown> };

declare global { interface Window { ethereum?: EthereumProvider } }

export function Faucet() {
  const [session, setSession] = useState<XSession>({ authenticated: false });
  const [sessionLoading, setSessionLoading] = useState(true);
  const [address, setAddress] = useState("");
  const [turnstileToken, setTurnstileToken] = useState("");
  const [resetKey, setResetKey] = useState(0);
  const [pending, setPending] = useState(false);
  const [message, setMessage] = useState<{ kind: "error" | "success"; text: string } | null>(null);

  useEffect(() => {
    fetch("/api/x/session", { cache: "no-store" })
      .then((response) => response.json())
      .then((value) => setSession(value.authenticated ? value : { authenticated: false }))
      .catch(() => setSession({ authenticated: false }))
      .finally(() => setSessionLoading(false));
  }, []);

  const connectX = async () => {
    setMessage(null);
    try {
      const response = await fetch("/api/x/start?returnTo=/", { cache: "no-store" });
      const result = await response.json();
      if (!response.ok || !result.authorizationUrl) throw new Error(result.error || "X verification is unavailable.");
      window.location.assign(result.authorizationUrl);
    } catch (reason) {
      setMessage({ kind: "error", text: reason instanceof Error ? reason.message : "Could not connect X." });
    }
  };

  const disconnectX = async () => {
    await fetch("/api/x/session", { method: "DELETE" }).catch(() => undefined);
    setSession({ authenticated: false });
    setMessage(null);
  };

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
    () => Boolean(session.authenticated && session.handle && isEthereumAddress(address) && turnstileToken && !pending),
    [address, pending, session, turnstileToken],
  );

  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (!ready || !session.handle) return;
    setPending(true);
    setMessage(null);
    try {
      const response = await fetch("/api/faucet", {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ xHandle: session.handle, recipientAddress: address.trim(), turnstileToken }),
      });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || result.errors?.[0]?.message || "The request failed.");
      const parsed = readFaucetResult(result);
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
          <span className="field__label" id="x-label">X account</span>
          <div className="field__control">
            {sessionLoading ? (
              <span className="skeleton">Checking…</span>
            ) : session.authenticated && session.handle ? (
              <div className="field__value">
                <strong>@{session.handle}</strong>
                <button type="button" className="text-button" onClick={disconnectX}>Disconnect</button>
              </div>
            ) : (
              <button type="button" className="btn btn--ghost" onClick={connectX} aria-labelledby="x-label">Connect X</button>
            )}
          </div>
        </div>

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

      {message && <div className={`form-message ${message.kind}`} role="status">{message.text}</div>}

      <div className="faucet__actions">
        <span className="meta">One claim · verified X account</span>
        <button className="btn btn--primary btn--send" type="submit" disabled={!ready}>
          {pending ? "Sending…" : "Send tokens"}<span aria-hidden="true">→</span>
        </button>
      </div>
    </form>
  );
}
