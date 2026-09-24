"use client";

import { useEffect, useId, useRef } from "react";

declare global {
  interface Window {
    turnstile?: {
      render: (element: HTMLElement, options: Record<string, unknown>) => string;
      remove: (id: string) => void;
      reset: (id: string) => void;
    };
  }
}

type Props = { onToken: (token: string) => void; resetKey: number };

export function Turnstile({ onToken, resetKey }: Props) {
  const siteKey = process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY;
  const elementRef = useRef<HTMLDivElement>(null);
  const widgetRef = useRef<string | null>(null);
  const id = useId().replace(/:/g, "");

  useEffect(() => {
    if (!siteKey) return;
    const render = () => {
      if (!elementRef.current || !window.turnstile || widgetRef.current) return;
      widgetRef.current = window.turnstile.render(elementRef.current, {
        sitekey: siteKey,
        theme: "light",
        size: "flexible",
        callback: (token: string) => onToken(token),
        "expired-callback": () => onToken(""),
        "error-callback": () => onToken(""),
      });
    };
    const existing = document.querySelector<HTMLScriptElement>('script[data-florin-turnstile]');
    if (existing) { if (window.turnstile) render(); else existing.addEventListener("load", render, { once: true }); }
    else {
      const script = document.createElement("script");
      script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";
      script.async = true;
      script.defer = true;
      script.dataset.florinTurnstile = "true";
      script.addEventListener("load", render, { once: true });
      document.head.appendChild(script);
    }
    return () => {
      if (widgetRef.current && window.turnstile) window.turnstile.remove(widgetRef.current);
      widgetRef.current = null;
    };
  }, [id, onToken, siteKey]);

  useEffect(() => {
    if (widgetRef.current && window.turnstile) window.turnstile.reset(widgetRef.current);
  }, [resetKey]);

  if (!siteKey) return <p className="config-note">Turnstile site key is not configured.</p>;
  return <div id={`turnstile-${id}`} ref={elementRef} className="turnstile" />;
}
