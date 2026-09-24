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
    <form className="faucet-card" onSubmit={submit}>
      <div className="step">
        <div className="step-number">01</div>
        <div className="step-body">
          <label>Verify your X account</label>
          <p>One account, one verified identity.</p>
          {sessionLoading ? (
            <div className="skeleton">Checking connection…</div>
          ) : session.authenticated && session.handle ? (
            <div className="connected-row">
              <span className="x-avatar">𝕏</span>
              <span><strong>@{session.handle}</strong><small>Verified</small></span>
              <button type="button" className="text-button" onClick={disconnectX}>Disconnect</button>
            </div>
          ) : (
            <button type="button" className="pill-button dark" onClick={connectX}><span>𝕏</span> Connect X</button>
          )}
        </div>
      </div>

      <div className="step">
        <div className="step-number">02</div>
        <div className="step-body">
          <label htmlFor="address">Choose a recipient</label>
          <p>Connect a wallet or paste any EVM address.</p>
          <div className="address-control">
            <input
              id="address"
              name="address"
              value={address}
              onChange={(event) => setAddress(event.target.value)}
              placeholder="0x0000…0000"
              autoComplete="off"
              spellCheck={false}
              aria-invalid={Boolean(address && !isEthereumAddress(address))}
            />
            <button type="button" onClick={connectWallet}>Connect wallet</button>
          </div>
          {address && !isEthereumAddress(address) && <small className="field-error">Enter a 42-character EVM address.</small>}
        </div>
      </div>

      <div className="step last-step">
        <div className="step-number">03</div>
        <div className="step-body">
          <label>Confirm you’re human</label>
          <p>A quick Cloudflare check helps keep the faucet open to everyone.</p>
          <Turnstile onToken={onToken} resetKey={resetKey} />
        </div>
      </div>

      {message && <div className={`form-message ${message.kind}`} role="status">{message.text}</div>}

      <div className="submit-row">
        <span>Available once every 24 hours</span>
        <button className="submit-button" type="submit" disabled={!ready}>
          {pending ? "Requesting…" : "Request funds"}<b aria-hidden="true">→</b>
        </button>
      </div>
    </form>
  );
}
