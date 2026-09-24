import { Suspense } from "react";
import XCallback from "@/components/x-callback";

export default function XCallbackPage() {
  return (
    <Suspense fallback={
      <main className="callback-page">
        <span className="brand-mark callback-mark" aria-hidden="true" />
        <p className="eyebrow">Florin / X verification</p>
        <h1>Verifying your account…</h1>
      </main>
    }>
      <XCallback />
    </Suspense>
  );
}
