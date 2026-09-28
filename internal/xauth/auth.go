package xauth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	transactionCookieName = "__Host-florin_x_oauth"
	sessionCookieName     = "__Host-florin_x_session"
	transactionPurpose    = "florin-x-oauth-transaction-v1"
	sessionPurpose        = "florin-x-session-v1"
	maxResponseBytes      = 1 << 20
	maxFollowingPages     = 25
	followingCheckTimeout = 20 * time.Second
)

var ErrUnauthenticated = errors.New("X session is missing or invalid")

type Config struct {
	ClientID         string
	ClientSecret     string
	SessionSecret    string
	AllowedRedirects []string
	RequiredUsername string
	RequiredUserID   string
	AuthorizeURL     string
	TokenURL         string
	APIBaseURL       string
	SessionTTL       time.Duration
	TransactionTTL   time.Duration
	HTTPClient       *http.Client
	Now              func() time.Time
	Random           io.Reader
}

type Auth struct {
	config           Config
	cookies          cipher.AEAD
	allowedRedirects map[string]struct{}
	stateMu          sync.Mutex
	usedTransactions map[string]int64
	revokedSessions  map[string]int64
}

type Identity struct {
	SessionID      string
	UserID         string
	Handle         string
	AccessToken    string
	RequiredUserID string
	ExpiresAt      time.Time
}

type identityContextKey struct{}

// IdentityFromContext returns the backend-verified X identity attached by Middleware.
func IdentityFromContext(ctx context.Context) (Identity, bool) {
	identity, ok := ctx.Value(identityContextKey{}).(Identity)
	return identity, ok
}

// ContextWithIdentity attaches a verified identity. HTTP requests should use Middleware;
// this helper also makes non-HTTP claim authorization straightforward to test.
func ContextWithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityContextKey{}, identity)
}

// Middleware attaches a valid X session to the request context. It deliberately
// leaves unauthenticated requests in place so GraphQL can return a field error.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if identity, err := a.Identity(r); err == nil {
			r = r.WithContext(context.WithValue(r.Context(), identityContextKey{}, identity))
		} else {
			slog.Error("failed to attach X identity to request", "err", err)
		}
		next.ServeHTTP(w, r)
	})
}

type transaction struct {
	State       string `json:"state"`
	Verifier    string `json:"verifier"`
	RedirectURI string `json:"redirect_uri"`
	ReturnTo    string `json:"return_to"`
	ExpiresAt   int64  `json:"expires_at"`
}

type session struct {
	SessionID      string `json:"session_id"`
	UserID         string `json:"user_id"`
	Handle         string `json:"handle"`
	AccessToken    string `json:"access_token"`
	RequiredUserID string `json:"required_user_id"`
	ExpiresAt      int64  `json:"expires_at"`
}

type callbackRequest struct {
	Code        string `json:"code"`
	State       string `json:"state"`
	RedirectURI string `json:"redirectUri"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type xUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type userResponse struct {
	Data xUser `json:"data"`
}

type followingResponse struct {
	Data []xUser `json:"data"`
	Meta struct {
		NextToken string `json:"next_token"`
	} `json:"meta"`
}

func New(config Config) (*Auth, error) {
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.RequiredUsername = strings.TrimPrefix(strings.TrimSpace(config.RequiredUsername), "@")
	config.RequiredUserID = strings.TrimSpace(config.RequiredUserID)
	if config.ClientID == "" {
		slog.Error("X client ID is required")
		return nil, errors.New("X client ID is required")
	}
	if len(config.SessionSecret) < 32 {
		slog.Error("X session secret must contain at least 32 characters")
		return nil, errors.New("X session secret must contain at least 32 characters")
	}
	if len(config.AllowedRedirects) == 0 {
		slog.Error("at least one X redirect URI is required")
		return nil, errors.New("at least one X redirect URI is required")
	}
	if config.RequiredUsername == "" && config.RequiredUserID == "" {
		slog.Error("required X username or user ID is required")
		return nil, errors.New("required X username or user ID is required")
	}
	if config.AuthorizeURL == "" {
		config.AuthorizeURL = "https://x.com/i/oauth2/authorize"
	}
	if config.TokenURL == "" {
		config.TokenURL = "https://api.x.com/2/oauth2/token"
	}
	if config.APIBaseURL == "" {
		config.APIBaseURL = "https://api.x.com/2"
	}
	config.APIBaseURL = strings.TrimRight(config.APIBaseURL, "/")
	if config.SessionTTL <= 0 {
		config.SessionTTL = 24 * time.Hour
	}
	if config.TransactionTTL <= 0 {
		config.TransactionTTL = 10 * time.Minute
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	if config.Random == nil {
		config.Random = rand.Reader
	}

	allowed := make(map[string]struct{}, len(config.AllowedRedirects))
	for _, raw := range config.AllowedRedirects {
		redirect := strings.TrimSpace(raw)
		parsed, err := url.Parse(redirect)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" || parsed.Fragment != "" {
			slog.Error("invalid allowed X redirect URI", "err", err)
			return nil, fmt.Errorf("invalid allowed X redirect URI %q", redirect)
		}
		allowed[redirect] = struct{}{}
	}

	instanceNonce := make([]byte, 32)
	if _, err := io.ReadFull(config.Random, instanceNonce); err != nil {
		slog.Error("failed to generate cookie instance nonce", "err", err)
		return nil, fmt.Errorf("generate cookie instance nonce: %w", err)
	}
	keyMaterial := append([]byte(config.SessionSecret), instanceNonce...)
	key := sha256.Sum256(keyMaterial)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		slog.Error("failed to create cookie cipher", "err", err)
		return nil, fmt.Errorf("create cookie cipher: %w", err)
	}
	cookies, err := cipher.NewGCM(block)
	if err != nil {
		slog.Error("failed to create cookie AEAD", "err", err)
		return nil, fmt.Errorf("create cookie AEAD: %w", err)
	}
	return &Auth{
		config:           config,
		cookies:          cookies,
		allowedRedirects: allowed,
		usedTransactions: make(map[string]int64),
		revokedSessions:  make(map[string]int64),
	}, nil
}

func (a *Auth) StartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	redirectURI := r.URL.Query().Get("redirectUri")
	if _, ok := a.allowedRedirects[redirectURI]; !ok {
		writeError(w, http.StatusBadRequest, "Invalid X callback URL.")
		return
	}
	state, err := a.randomToken(32)
	if err != nil {
		slog.Error("failed to generate X OAuth state", "err", err)
		writeError(w, http.StatusInternalServerError, "Could not start X verification.")
		return
	}
	verifier, err := a.randomToken(32)
	if err != nil {
		slog.Error("failed to generate X OAuth PKCE verifier", "err", err)
		writeError(w, http.StatusInternalServerError, "Could not start X verification.")
		return
	}
	challengeBytes := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(challengeBytes[:])
	tx := transaction{
		State:       state,
		Verifier:    verifier,
		RedirectURI: redirectURI,
		ReturnTo:    safeReturnTo(r.URL.Query().Get("returnTo")),
		ExpiresAt:   a.config.Now().Add(a.config.TransactionTTL).Unix(),
	}
	cookieValue, err := a.seal(transactionPurpose, tx)
	if err != nil {
		slog.Error("failed to seal X OAuth transaction cookie", "err", err)
		writeError(w, http.StatusInternalServerError, "Could not start X verification.")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     transactionCookieName,
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(a.config.TransactionTTL.Seconds()),
	})

	authorize, err := url.Parse(a.config.AuthorizeURL)
	if err != nil {
		slog.Error("failed to parse X authorization URL", "err", err)
		writeError(w, http.StatusInternalServerError, "X verification is misconfigured.")
		return
	}
	query := authorize.Query()
	query.Set("response_type", "code")
	query.Set("client_id", a.config.ClientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", "tweet.read users.read follows.read")
	query.Set("state", state)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	authorize.RawQuery = query.Encode()
	writeJSON(w, http.StatusOK, map[string]string{"authorizationUrl": authorize.String()})
}

func (a *Auth) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
		return
	}
	cookie, err := r.Cookie(transactionCookieName)
	if err != nil {
		slog.Error("X OAuth transaction cookie is missing", "err", err)
		writeError(w, http.StatusBadRequest, "X verification has expired. Please start again.")
		return
	}
	var tx transaction
	openErr := a.open(transactionPurpose, cookie.Value, &tx)
	if openErr != nil || tx.ExpiresAt <= a.config.Now().Unix() {
		slog.Error("X OAuth transaction cookie is invalid or expired", "err", openErr)
		a.clearCookie(w, transactionCookieName)
		writeError(w, http.StatusBadRequest, "X verification has expired. Please start again.")
		return
	}

	var callback callbackRequest
	if err := decodeJSON(r.Body, &callback); err != nil {
		slog.Error("failed to decode X OAuth callback", "err", err)
		writeError(w, http.StatusBadRequest, "Invalid X callback.")
		return
	}
	if callback.Code == "" || callback.State == "" || callback.RedirectURI == "" ||
		subtle.ConstantTimeCompare([]byte(callback.State), []byte(tx.State)) != 1 ||
		callback.RedirectURI != tx.RedirectURI {
		writeError(w, http.StatusBadRequest, "Invalid X callback.")
		return
	}
	if _, ok := a.allowedRedirects[callback.RedirectURI]; !ok {
		writeError(w, http.StatusBadRequest, "Invalid X callback URL.")
		return
	}
	if !a.consumeTransaction(tx.State, tx.ExpiresAt) {
		a.clearCookie(w, transactionCookieName)
		writeError(w, http.StatusBadRequest, "X verification was already used. Please start again.")
		return
	}
	a.clearCookie(w, transactionCookieName)

	token, err := a.exchangeCode(r.Context(), callback.Code, tx.Verifier, callback.RedirectURI)
	if err != nil {
		slog.Error("failed to exchange X OAuth code", "err", err)
		writeError(w, http.StatusBadGateway, "X did not accept the verification callback. Please try again.")
		return
	}
	user, err := a.currentUser(r.Context(), token.AccessToken)
	if err != nil {
		slog.Error("failed to read current X user", "err", err)
		writeError(w, http.StatusBadGateway, "Could not read your X account. Please try again.")
		return
	}
	requiredID := a.config.RequiredUserID
	if requiredID == "" {
		required, lookupErr := a.userByUsername(r.Context(), token.AccessToken, a.config.RequiredUsername)
		if lookupErr != nil {
			slog.Error("failed to resolve required X user", "err", lookupErr)
			writeError(w, http.StatusBadGateway, "Could not verify the required X account. Please try again.")
			return
		}
		requiredID = required.ID
	}
	follows, err := a.follows(r.Context(), token.AccessToken, user.ID, requiredID)
	if err != nil {
		slog.Error("failed to check X follow relationship", "err", err)
		writeError(w, http.StatusBadGateway, "Could not check who you follow on X. Please try again.")
		return
	}
	if !follows {
		writeError(w, http.StatusForbidden, "Follow @"+a.config.RequiredUsername+" on X, then try again.")
		return
	}

	expiresAt := a.config.Now().Add(a.config.SessionTTL)
	if token.ExpiresIn > 0 {
		tokenExpiry := a.config.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
		if tokenExpiry.Before(expiresAt) {
			expiresAt = tokenExpiry
		}
	}
	sessionID, err := a.randomToken(32)
	if err != nil {
		slog.Error("failed to generate X session ID", "err", err)
		writeError(w, http.StatusInternalServerError, "Could not create an X session.")
		return
	}
	payload := session{
		SessionID:      sessionID,
		UserID:         user.ID,
		Handle:         user.Username,
		AccessToken:    token.AccessToken,
		RequiredUserID: requiredID,
		ExpiresAt:      expiresAt.Unix(),
	}
	value, err := a.seal(sessionPurpose, payload)
	if err != nil {
		slog.Error("failed to seal X session cookie", "err", err)
		writeError(w, http.StatusInternalServerError, "Could not create an X session.")
		return
	}
	maxAge := int(time.Until(expiresAt).Seconds())
	if a.config.Now != nil {
		maxAge = int(expiresAt.Sub(a.config.Now()).Seconds())
	}
	if maxAge < 1 {
		writeError(w, http.StatusBadGateway, "X returned an expired session. Please try again.")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	})
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"handle":        user.Username,
		"returnTo":      tx.ReturnTo,
	})
}

func (a *Auth) SessionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		if identity, err := a.Identity(r); err == nil {
			a.revokeSession(identity.SessionID, identity.ExpiresAt.Unix())
		} else {
			slog.Error("failed to read X identity during logout", "err", err)
		}
		a.clearCookie(w, sessionCookieName)
		writeJSON(w, http.StatusOK, map[string]bool{"authenticated": false})
	case http.MethodGet:
		identity, err := a.Identity(r)
		if err != nil {
			slog.Error("failed to read X session", "err", err)
			writeJSON(w, http.StatusOK, map[string]bool{"authenticated": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"authenticated": true, "handle": identity.Handle})
	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodDelete)
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed.")
	}
}

func (a *Auth) Identity(r *http.Request) (Identity, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		slog.Error("X session cookie is missing", "err", err)
		return Identity{}, ErrUnauthenticated
	}
	var stored session
	openErr := a.open(sessionPurpose, cookie.Value, &stored)
	if openErr != nil || stored.ExpiresAt <= a.config.Now().Unix() ||
		stored.SessionID == "" || stored.UserID == "" || stored.Handle == "" || stored.AccessToken == "" || stored.RequiredUserID == "" ||
		a.sessionIsRevoked(stored.SessionID) {
		slog.Error("X session cookie is invalid, expired, incomplete, or revoked", "err", openErr)
		return Identity{}, ErrUnauthenticated
	}
	return Identity{
		SessionID:      stored.SessionID,
		UserID:         stored.UserID,
		Handle:         stored.Handle,
		AccessToken:    stored.AccessToken,
		RequiredUserID: stored.RequiredUserID,
		ExpiresAt:      time.Unix(stored.ExpiresAt, 0),
	}, nil
}

func (a *Auth) VerifyFollowing(ctx context.Context, identity Identity) (bool, error) {
	if identity.UserID == "" || identity.AccessToken == "" || identity.RequiredUserID == "" {
		slog.Error("cannot verify X following for an incomplete identity")
		return false, ErrUnauthenticated
	}
	return a.follows(ctx, identity.AccessToken, identity.UserID, identity.RequiredUserID)
}

func (a *Auth) exchangeCode(ctx context.Context, code, verifier, redirectURI string) (tokenResponse, error) {
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
	if a.config.ClientSecret == "" {
		form.Set("client_id", a.config.ClientID)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.config.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		slog.Error("failed to create X token request", "err", err)
		return tokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if a.config.ClientSecret != "" {
		req.SetBasicAuth(a.config.ClientID, a.config.ClientSecret)
	}
	var token tokenResponse
	if err := a.doJSON(req, &token); err != nil {
		slog.Error("X token request failed", "err", err)
		return token, err
	}
	if token.AccessToken == "" || !strings.EqualFold(token.TokenType, "bearer") {
		slog.Error("X token response is incomplete")
		return token, errors.New("X token response is incomplete")
	}
	return token, nil
}

func (a *Auth) currentUser(ctx context.Context, accessToken string) (xUser, error) {
	var response userResponse
	if err := a.apiGET(ctx, accessToken, "/users/me", nil, &response); err != nil {
		slog.Error("failed to fetch current X user", "err", err)
		return xUser{}, err
	}
	if response.Data.ID == "" || response.Data.Username == "" {
		slog.Error("X user response is incomplete")
		return xUser{}, errors.New("X user response is incomplete")
	}
	return response.Data, nil
}

func (a *Auth) userByUsername(ctx context.Context, accessToken, username string) (xUser, error) {
	var response userResponse
	endpoint := "/users/by/username/" + url.PathEscape(username)
	if err := a.apiGET(ctx, accessToken, endpoint, nil, &response); err != nil {
		slog.Error("failed to fetch required X user", "err", err)
		return xUser{}, err
	}
	if response.Data.ID == "" {
		slog.Error("required X user was not found")
		return xUser{}, errors.New("required X user was not found")
	}
	return response.Data, nil
}

func (a *Auth) follows(ctx context.Context, accessToken, sourceUserID, targetUserID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, followingCheckTimeout)
	defer cancel()

	paginationToken := ""
	seen := make(map[string]struct{})
	for pageNumber := 0; pageNumber < maxFollowingPages; pageNumber++ {
		query := url.Values{"max_results": {"1000"}}
		if paginationToken != "" {
			query.Set("pagination_token", paginationToken)
		}
		var response followingResponse
		endpoint := path.Join("/users", url.PathEscape(sourceUserID), "following")
		if err := a.apiGET(ctx, accessToken, endpoint, query, &response); err != nil {
			slog.Error("failed to fetch X following page", "err", err, "page", pageNumber+1)
			return false, err
		}
		for _, followed := range response.Data {
			if subtle.ConstantTimeCompare([]byte(followed.ID), []byte(targetUserID)) == 1 {
				return true, nil
			}
		}
		next := response.Meta.NextToken
		if next == "" {
			return false, nil
		}
		if _, duplicate := seen[next]; duplicate {
			slog.Error("X returned a pagination cycle")
			return false, errors.New("X returned a pagination cycle")
		}
		seen[next] = struct{}{}
		paginationToken = next
	}
	slog.Error("X following pagination exceeded its limit", "limit", maxFollowingPages)
	return false, errors.New("X following pagination exceeded its limit")
}

func (a *Auth) apiGET(ctx context.Context, accessToken, endpoint string, query url.Values, destination any) error {
	requestURL := a.config.APIBaseURL + endpoint
	if len(query) > 0 {
		requestURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		slog.Error("failed to create X API request", "err", err)
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	return a.doJSON(req, destination)
}

func (a *Auth) doJSON(req *http.Request, destination any) error {
	response, err := a.config.HTTPClient.Do(req)
	if err != nil {
		slog.Error("X HTTP request failed", "err", err, "method", req.Method, "url", req.URL.Redacted())
		return err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			slog.Error("failed to close X HTTP response body", "err", err)
		}
	}()
	body := io.LimitReader(response.Body, maxResponseBytes+1)
	data, err := io.ReadAll(body)
	if err != nil {
		slog.Error("failed to read X HTTP response", "err", err)
		return err
	}
	if len(data) > maxResponseBytes {
		slog.Error("X response is too large", "limit", maxResponseBytes)
		return errors.New("X response is too large")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		slog.Error("X returned an unsuccessful HTTP status", "status", response.StatusCode)
		return fmt.Errorf("X returned HTTP %d", response.StatusCode)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		slog.Error("failed to decode X response", "err", err)
		return fmt.Errorf("decode X response: %w", err)
	}
	return nil
}

func (a *Auth) randomToken(size int) (string, error) {
	data := make([]byte, size)
	if _, err := io.ReadFull(a.config.Random, data); err != nil {
		slog.Error("failed to generate random token", "err", err)
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func (a *Auth) seal(purpose string, value any) (string, error) {
	plaintext, err := json.Marshal(value)
	if err != nil {
		slog.Error("failed to encode encrypted cookie payload", "err", err)
		return "", err
	}
	nonce := make([]byte, a.cookies.NonceSize())
	if _, err := io.ReadFull(a.config.Random, nonce); err != nil {
		slog.Error("failed to generate encrypted cookie nonce", "err", err)
		return "", err
	}
	sealed := a.cookies.Seal(nonce, nonce, plaintext, []byte(purpose))
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (a *Auth) open(purpose, value string, destination any) error {
	sealed, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(sealed) < a.cookies.NonceSize() {
		slog.Error("failed to decode encrypted cookie", "err", err)
		return ErrUnauthenticated
	}
	nonce := sealed[:a.cookies.NonceSize()]
	plaintext, err := a.cookies.Open(nil, nonce, sealed[a.cookies.NonceSize():], []byte(purpose))
	if err != nil {
		slog.Error("failed to open encrypted cookie", "err", err)
		return ErrUnauthenticated
	}
	if err := json.Unmarshal(plaintext, destination); err != nil {
		slog.Error("failed to decode encrypted cookie payload", "err", err)
		return ErrUnauthenticated
	}
	return nil
}

func (a *Auth) consumeTransaction(state string, expiresAt int64) bool {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.removeExpiredStateLocked()
	if _, used := a.usedTransactions[state]; used {
		return false
	}
	a.usedTransactions[state] = expiresAt
	return true
}

func (a *Auth) revokeSession(sessionID string, expiresAt int64) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.removeExpiredStateLocked()
	a.revokedSessions[sessionID] = expiresAt
}

func (a *Auth) sessionIsRevoked(sessionID string) bool {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.removeExpiredStateLocked()
	_, revoked := a.revokedSessions[sessionID]
	return revoked
}

func (a *Auth) removeExpiredStateLocked() {
	now := a.config.Now().Unix()
	for state, expiresAt := range a.usedTransactions {
		if expiresAt <= now {
			delete(a.usedTransactions, state)
		}
	}
	for sessionID, expiresAt := range a.revokedSessions {
		if expiresAt <= now {
			delete(a.revokedSessions, sessionID)
		}
	}
}

func (a *Auth) clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func safeReturnTo(value string) string {
	if value == "" || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") {
		slog.Error("invalid OAuth return path")
		return "/"
	}
	decoded, err := url.PathUnescape(value)
	if err != nil || strings.HasPrefix(decoded, "//") || strings.ContainsRune(decoded, '\\') {
		slog.Error("failed to validate OAuth return path", "err", err)
		return "/"
	}
	for _, character := range decoded {
		if character < 0x20 || character == 0x7f {
			slog.Error("OAuth return path contains a control character")
			return "/"
		}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		slog.Error("failed to parse OAuth return path", "err", err)
		return "/"
	}
	return value
}

func decodeJSON(reader io.Reader, destination any) error {
	data, err := io.ReadAll(io.LimitReader(reader, (32<<10)+1))
	if err != nil {
		slog.Error("failed to read JSON request", "err", err)
		return err
	}
	if len(data) > 32<<10 {
		slog.Error("JSON request is too large", "limit", 32<<10)
		return errors.New("JSON request is too large")
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		slog.Error("failed to decode JSON request", "err", err)
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		slog.Error("JSON request must contain exactly one value", "err", err)
		return errors.New("JSON request must contain exactly one value")
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	slog.Error("X authentication request failed", "status", status, "error", message)
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to write JSON response", "err", err, "status", status)
	}
}
