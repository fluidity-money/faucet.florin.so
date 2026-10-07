# Optional X authentication

The Go service retains an optional X OAuth 2.0 Authorization Code flow with PKCE. Faucet claims do not require an X session, and the OAuth flow does not check social relationships.

The browser never receives the X client secret or bearer token. OAuth transaction data and the short-lived X access token are encrypted with AES-GCM inside purpose-bound HttpOnly cookies. OAuth state is consumed in a concurrency-safe in-memory replay cache before code exchange, and logout revokes the session ID until its expiry.

The replay/revocation cache is process-local because the service currently runs as one Unix-socket process. Each process start mixes a fresh random generation nonce into the cookie-encryption key, so a restart invalidates prior sessions and in-progress OAuth transactions.

## X developer-console setup

Create an OAuth 2.0 app in the X developer portal and configure it as a Web App / confidential client.

Grant these scopes:

```text
tweet.read users.read
```

Configure the callback URLs needed by each environment. The callback string must exactly match both the X app configuration and `SPN_X_REDIRECT_URIS`.

## Go service environment

```text
SPN_X_CLIENT_ID=[REDACTED]
SPN_X_CLIENT_SECRET=[REDACTED]
SPN_X_SESSION_SECRET=[REDACTED]
SPN_X_REDIRECT_URIS=http://localhost:3000/auth/x/callback,https://faucet.florin.so/auth/x/callback
```

`SPN_X_CLIENT_SECRET` may be omitted only if the X app is intentionally configured as a public PKCE client. A normal server-side Web App should set it.

The existing backend variables are still required:

```text
SPN_TURNSTILE_SECRET=***
SPN_TURNSTILE_HOSTNAMES=faucet.florin.so
SPN_ROBINHOOD_URL=[REDACTED]
SPN_TIMESCALE=[REDACTED]
```

Keep all of these variables server-side. The public Turnstile site key is embedded in `components/turnstile.tsx`.

## HTTP contract

All endpoints live on `FAUCET_GRAPH_URL`. Errors use `{ "error": "human-readable message" }`.

### `GET /auth/x/start`

Query parameters:

- `redirectUri`: exact allowlisted callback URI.
- `returnTo`: local frontend path, normally `/`.

The handler creates random OAuth state and a PKCE S256 verifier/challenge, stores the transaction in an encrypted ten-minute HttpOnly cookie, and returns an `authorizationUrl`.

### `POST /auth/x/callback`

The callback exchanges the code with the stored PKCE verifier, calls `GET /2/users/me`, creates an encrypted `__Host-florin_x_session` cookie bounded by the access-token lifetime, and returns the authenticated handle.

### `GET /auth/x/session`

Returns either an authenticated session and handle or `{ "authenticated": false }`.

### `DELETE /auth/x/session`

Expires the session cookie and returns `{ "authenticated": false }`.

## Faucet claims

The claim mutation is independent of X:

```graphql
requestToken(recipientAddress: String!, turnstileToken: String!): Boolean!
```

The faucet stores `NULL` in `x_handle` for these claims. Address and time-window abuse controls remain unchanged.

## Deployment notes

- OAuth and session cookies use the `__Host-` prefix with `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, and no `Domain`.
- Do not log callback codes, access tokens, session cookies, or `SPN_X_SESSION_SECRET`.
