# X authentication and follow verification

The Go service now implements the X OAuth 2.0 Authorization Code flow with PKCE and requires each claimant to follow `@florinprotocol`.

The browser never receives the X client secret or bearer token. OAuth transaction data and the short-lived X access token are encrypted with AES-GCM inside purpose-bound HttpOnly cookies. OAuth state is also consumed in a concurrency-safe in-memory replay cache before code exchange, and logout revokes the session ID until its expiry. The callback checks the follow relationship before establishing a session, and the GraphQL claim path checks the relationship again immediately before a claim.

The replay/revocation cache is process-local because the service currently runs as one Unix-socket process. If the service is later replicated, move those two expiry maps to shared storage such as Postgres or Redis before running more than one instance.

## X developer-console setup

Create an OAuth 2.0 app in the X developer portal and configure it as a Web App / confidential client.

Grant these scopes:

```text
tweet.read users.read follows.read
```

`follows.read` is required for `GET /2/users/:id/following`. No write or offline scope is requested.

Configure these callback URLs exactly as needed by your environments:

```text
http://localhost:3000/auth/x/callback
https://faucet.florin.so/auth/x/callback
```

The callback string must exactly match both the X app configuration and `SPN_X_REDIRECT_URIS`.

## Go service environment

```text
SPN_X_CLIENT_ID=your-oauth2-client-id
SPN_X_CLIENT_SECRET=your-oauth2-client-secret
SPN_X_SESSION_SECRET=a-random-secret-containing-at-least-32-characters
SPN_X_REDIRECT_URIS=http://localhost:3000/auth/x/callback,https://faucet.florin.so/auth/x/callback
SPN_X_REQUIRED_USERNAME=florinprotocol
SPN_X_REQUIRED_USER_ID=the-immutable-x-user-id-for-florinprotocol
```

Generate the cookie-encryption secret with, for example:

```sh
openssl rand -base64 48
```

`SPN_X_REQUIRED_USER_ID` is optional, but production should set it. If omitted, the callback resolves `florinprotocol` through `GET /2/users/by/username/florinprotocol`. Pinning the immutable ID prevents a renamed or transferred username from changing the account users are required to follow.

`SPN_X_CLIENT_SECRET` may be omitted only if the X app is intentionally configured as a public PKCE client. A normal server-side Web App should set it.

The existing backend variables are still required:

```text
SPN_TURNSTILE_SECRET=your-cloudflare-turnstile-secret
SPN_ROBINHOOD_URL=your-evm-rpc-url
SPN_TIMESCALE=your-postgres-dsn
```

Keep all of these variables server-side. Only `NEXT_PUBLIC_TURNSTILE_SITE_KEY` belongs in the browser bundle.

## Implemented HTTP contract

All endpoints live on `FAUCET_GRAPH_URL`. Errors use `{ "error": "human-readable message" }`.

### `GET /auth/x/start`

Query parameters:

- `redirectUri`: exact allowlisted callback URI.
- `returnTo`: local frontend path, normally `/`.

The handler creates random OAuth state and a PKCE S256 verifier/challenge, stores the transaction in an encrypted ten-minute HttpOnly cookie, and returns:

```json
{ "authorizationUrl": "https://x.com/i/oauth2/authorize?..." }
```

### `POST /auth/x/callback`

The Next callback page sends:

```json
{
  "code": "one-time-code-from-x",
  "state": "state-from-x",
  "redirectUri": "https://faucet.florin.so/auth/x/callback"
}
```

The handler:

1. decrypts and expires the OAuth transaction cookie;
2. compares state in constant time and requires the exact stored callback URI;
3. exchanges the code with the stored PKCE verifier;
4. calls `GET /2/users/me`;
5. resolves the required account ID if it was not pinned in configuration;
6. paginates `GET /2/users/:id/following` until it finds `@florinprotocol` or exhausts the result set;
7. refuses non-followers with HTTP 403;
8. creates an encrypted `florin_x_session` cookie bounded by the X access-token lifetime; and
9. returns the verified handle.

Successful response:

```json
{ "authenticated": true, "handle": "alice", "returnTo": "/" }
```

### `GET /auth/x/session`

Returns either:

```json
{ "authenticated": true, "handle": "alice" }
```

or:

```json
{ "authenticated": false }
```

### `DELETE /auth/x/session`

Expires the session cookie and returns:

```json
{ "authenticated": false }
```

## GraphQL claim enforcement

The schema remains unchanged:

```graphql
requestToken(xHandle: String!, recipientAddress: String!, turnstileToken: String!): Boolean!
```

The HTTP middleware decrypts the session and places the immutable X identity in the GraphQL request context. `authorizeXClaim` then:

1. rejects missing or invalid sessions;
2. rejects a submitted handle that does not match the verified account; and
3. queries X again to ensure the user still follows `@florinprotocol` at claim time.

The resolver deliberately fails closed if X is unavailable or rate-limits the check.

The current resolver still returns `token transfer is not implemented` after X authorization. Keep the `authorizeXClaim` call at the beginning when adding the transfer, Turnstile check, and once-per-day database transaction.

## Deployment notes

- The Next.js proxy must forward `Cookie` and relay every `Set-Cookie` response. It does this in `lib/backend.ts`.
- The proxy also forwards `X-Forwarded-Proto` and `X-Forwarded-Host`, allowing the Go service to issue/clear secure cookies correctly behind the Unix-socket proxy.
- X API access tiers and rate limits apply. The following-list endpoint is paginated with `max_results=1000`; claims fail closed rather than bypassing verification if X returns an error.
- Use the immutable X user ID for database abuse controls. Handles can change.
- Do not log callback codes, access tokens, session cookies, or `SPN_X_SESSION_SECRET`.
