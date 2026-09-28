import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./globals.css";

export const metadata: Metadata = {
  metadataBase: new URL("https://faucet.florin.so"),
  title: "Florin Faucet",
  description: "Get testnet funds for Florin on Robinhood Chain.",
  alternates: { canonical: "/" },
  openGraph: {
    type: "website",
    url: "/",
    siteName: "Florin",
    title: "Florin Faucet",
    description: "Get testnet funds for Florin on Robinhood Chain.",
    images: [
      {
        url: "/opengraph.png",
        width: 1200,
        height: 630,
        alt: "Florin Faucet — testnet funds for Florin on Robinhood Chain",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    site: "@FlorinProtocol",
    creator: "@FlorinProtocol",
    title: "Florin Faucet",
    description: "Get testnet funds for Florin on Robinhood Chain.",
    images: ["/opengraph.png"],
  },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
