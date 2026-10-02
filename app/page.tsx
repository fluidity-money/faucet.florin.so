import { Faucet } from "@/components/faucet";

export default function HomePage() {
  return (
    <main>
      <header className="site-header">
        <div className="site-header__row">
          <a className="brand" href="https://florin.so" aria-label="Florin home">
            <span className="brand__mark" aria-hidden="true" />
            <span>Florin</span>
          </a>
          <span className="network-pill">
            <span className="network-pill__dot" aria-hidden="true" />
            Robinhood Chain testnet
          </span>
        </div>
      </header>

      <section className="main">
        <div className="faucet">
          <p className="lead">
            Receive testnet SPY and 0.01 ETH every 2 hours.
          </p>
          <Faucet />
        </div>
      </section>

      <footer className="site-footer">
        <div className="site-footer__inner">
          <p className="site-footer__legal">FLORIN © 2026 · SUPERPOSITION</p>
          <nav className="site-footer__nav" aria-label="Footer">
            <a href="https://florin.so">Florin</a>
            <a href="https://superposition.so">Superposition</a>
          </nav>
        </div>
      </footer>
    </main>
  );
}