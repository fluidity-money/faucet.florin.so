"use client";

import { FormEvent, useCallback, useEffect, useMemo, useState } from "react";
import { readFaucetResult } from "@/lib/graphql";
import { isEthereumAddress } from "@/lib/validation";
import { Turnstile } from "./turnstile";

type XSession = { authenticated: boolean; handle?: string };
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

  const xConnected = Boolean(session.authenticated && session.handle);

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

  const clearXSession = async () => {
    setSession({ authenticated: false });
    setAddress("");
    setTurnstileToken("");
    setResetKey((value) => value + 1);
    await fetch("/api/x/session", { method: "DELETE" }).catch(() => undefined);
  };

  const disconnectX = async () => {
    await clearXSession();
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
    () => Boolean(xConnected && isEthereumAddress(address) && turnstileToken && !pending),
    [address, pending, turnstileToken, xConnected],
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
      const parsed = readFaucetResult(result);
      if (!parsed.ok && parsed.authenticationRequired) await clearXSession();
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
              <div className="x-connect">
                <a
                  className="x-connect__follow"
                  href="https://x.com/florinprotocol"
                  target="_blank"
                  rel="noreferrer"
                >
                  Follow <strong>@florinprotocol</strong> on X first <span aria-hidden="true">↗</span>
                </a>
                <button type="button" className="btn btn--ghost" onClick={connectX} aria-labelledby="x-label">Then connect X</button>
              </div>
            )}
          </div>
        </div>

        <div className={`field${xConnected ? "" : " field--locked"}`} aria-disabled={!xConnected}>
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
                disabled={!xConnected}
                aria-invalid={Boolean(address && !isEthereumAddress(address))}
                aria-labelledby="addr-label"
              />
              {address && isEthereumAddress(address) && <span className="addr__ok" aria-hidden="true">✓</span>}
              <button type="button" className="addr__btn" onClick={connectWallet} disabled={!xConnected}>Connect wallet</button>
            </div>
          </div>
          {address && !isEthereumAddress(address) && <small className="field-error">Enter a 42-character EVM address.</small>}
        </div>

        <div className={`field${xConnected ? "" : " field--locked"}`} aria-disabled={!xConnected}>
          <span className="field__label" id="human-label">Human check</span>
          <div className="field__control">
            {xConnected ? (
              <Turnstile onToken={onToken} resetKey={resetKey} />
            ) : (
              <span className="field-lock">Connect X to continue.</span>
            )}
          </div>
        </div>
      </div>

      {message?.kind === "success" ? (
        <ClaimSuccess />
      ) : message ? (
        <div className="form-message error" role="status">{message.text}</div>
      ) : null}

      <div className="faucet__actions">
        <span className="meta">Follow @florinprotocol · verified X account</span>
        <button className="btn btn--primary btn--send" type="submit" disabled={!ready}>
          {pending ? "Sending…" : "Send tokens"}<span aria-hidden="true">→</span>
        </button>
      </div>
    </form>
  );
}
