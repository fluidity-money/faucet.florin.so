"use client";

import { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { callbackPayload, safeReturnTo } from "@/lib/x-auth";

export default function XCallbackPage() {
  const params = useSearchParams();
  const router = useRouter();
  const [error, setError] = useState("");

  useEffect(() => {
    const oauthError = params.get("error");
    if (oauthError) { setError("X verification was cancelled or denied."); return; }
    let payload;
    try { payload = callbackPayload(params, `${window.location.origin}/auth/x/callback`); }
    catch (reason) { setError(reason instanceof Error ? reason.message : "Invalid callback."); return; }

    fetch("/api/x/callback", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify(payload),
    }).then(async (response) => {
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || "Could not verify your X account.");
      router.replace(safeReturnTo(result.returnTo));
    }).catch((reason) => setError(reason instanceof Error ? reason.message : "Could not verify your X account."));
  }, [params, router]);

  return (
    <main className="callback-page">
      <span className="brand-mark callback-mark" aria-hidden="true" />
      <p className="eyebrow">Florin / X verification</p>
      <h1>{error ? "Verification stopped." : "Verifying your account…"}</h1>
      {error && <a className="pill-button dark" href="/">Return to faucet</a>}
    </main>
  );
}
