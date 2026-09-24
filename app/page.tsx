import { Faucet } from "@/components/faucet";

export default function HomePage() {
  return (
    <main>
      <header className="site-header">
        <a className="brand" href="https://florin.so" aria-label="Florin home">
          <span className="brand-mark" aria-hidden="true" />
          <span>Florin</span>
        </a>
        <div className="header-meta">
          <span className="network-dot" aria-hidden="true" />
          Robinhood Chain testnet
        </div>
      </header>

      <section className="hero">
        <div className="hero-copy">
          <p className="eyebrow">Florin testnet faucet / 01</p>
          <h1>Fuel the first transaction.</h1>
          <p className="hero-lead">
            Verify a real X account, add your EVM address, and receive enough testnet funds to get moving.
          </p>
          <div className="hero-rule" aria-hidden="true">
            <span>01</span><i /><span>03</span>
          </div>
        </div>
        <div className="hero-art" aria-hidden="true">
          <div className="hero-shape" />
          <span className="hero-mark" />
          <span className="orbit orbit-one" />
          <span className="orbit orbit-two" />
          <p>VERIFY · FUND · BUILD</p>
        </div>
      </section>

      <section className="claim-section" id="claim">
        <div className="claim-intro">
          <p className="eyebrow">One claim per address, per day</p>
          <h2>Three checks.<br />Then you’re funded.</h2>
          <p>
            X verification and Turnstile keep the faucet useful for builders. Your X password is never shared with Florin.
          </p>
        </div>
        <Faucet />
      </section>

      <footer>
        <span>Florin is a Superposition property.</span>
        <div>
          <a href="https://florin.so">Florin</a>
          <a href="https://superposition.so">Superposition</a>
        </div>
      </footer>
    </main>
  );
}
