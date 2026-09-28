package xauth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestWriteErrorLogsResponse(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	recorder := httptest.NewRecorder()
	writeError(recorder, http.StatusBadGateway, "X request failed.")

	logged := logs.String()
	if !strings.Contains(logged, "level=ERROR") ||
		!strings.Contains(logged, "status=502") ||
		!strings.Contains(logged, `error="X request failed."`) {
		t.Fatalf("writeError log = %q", logged)
	}
}

const (
	testRedirect = "https://faucet.florin.so/auth/x/callback"
	testSecret   = "0123456789abcdef0123456789abcdef"
)

func testConfig(apiURL string) Config {
	return Config{
		ClientID:         "client-id",
		ClientSecret:     "client-secret",
		SessionSecret:    testSecret,
		AllowedRedirects: []string{testRedirect},
		RequiredUsername: "florinprotocol",
		RequiredUserID:   "999",
		AuthorizeURL:     apiURL + "/authorize",
		TokenURL:         apiURL + "/token",
		APIBaseURL:       apiURL,
		SessionTTL:       24 * time.Hour,
		TransactionTTL:   10 * time.Minute,
		Now:              func() time.Time { return time.Unix(1_800_000_000, 0) },
	}
}

func startLogin(t *testing.T, auth *Auth) (*http.Cookie, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/auth/x/start?redirectUri="+url.QueryEscape(testRedirect)+"&returnTo=%2F", nil)
	recorder := httptest.NewRecorder()
	auth.StartHandler(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("start status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var result struct {
		AuthorizationURL string `json:"authorizationUrl"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	authorizationURL, err := url.Parse(result.AuthorizationURL)
	if err != nil {
		t.Fatal(err)
	}
	query := authorizationURL.Query()
	if query.Get("scope") != "tweet.read users.read follows.read" {
		t.Fatalf("scope = %q", query.Get("scope"))
	}
	if query.Get("code_challenge_method") != "S256" || query.Get("code_challenge") == "" {
		t.Fatalf("missing PKCE parameters: %s", query.Encode())
	}
	if query.Get("state") == "" {
		t.Fatal("missing state")
	}
	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != transactionCookieName || !cookies[0].HttpOnly {
		t.Fatalf("unexpected cookies: %#v", cookies)
	}
	return cookies[0], query.Get("state")
}

func TestStartHandlerBuildsPKCEAuthorizationRequest(t *testing.T) {
	auth, err := New(testConfig("https://api.x.test"))
	if err != nil {
		t.Fatal(err)
	}
	startLogin(t, auth)
}

func TestStartHandlerRejectsUnlistedRedirect(t *testing.T) {
	auth, err := New(testConfig("https://api.x.test"))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/auth/x/start?redirectUri="+url.QueryEscape("https://evil.example/callback"), nil)
	recorder := httptest.NewRecorder()
	auth.StartHandler(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestCallbackVerifiesFollowAcrossAllPagesAndCreatesSession(t *testing.T) {
	var tokenCalls atomic.Int32
	var followingCalls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/token":
			tokenCalls.Add(1)
			if got, _, ok := r.BasicAuth(); !ok || got != "client-id" {
				t.Errorf("missing confidential-client auth")
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("code_verifier") == "" || r.Form.Get("redirect_uri") != testRedirect {
				t.Errorf("bad token form: %v", r.Form)
			}
			writeTestJSON(w, http.StatusOK, map[string]any{"access_token": "user-token", "token_type": "bearer", "expires_in": 7200})
		case r.URL.Path == "/users/me":
			assertBearer(t, r, "user-token")
			writeTestJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"id": "123", "username": "alice"}})
		case r.URL.Path == "/users/123/following":
			assertBearer(t, r, "user-token")
			followingCalls.Add(1)
			if r.URL.Query().Get("pagination_token") == "next-page" {
				writeTestJSON(w, http.StatusOK, map[string]any{"data": []map[string]string{{"id": "999", "username": "florinprotocol"}}, "meta": map[string]any{"result_count": 1}})
				return
			}
			writeTestJSON(w, http.StatusOK, map[string]any{"data": []map[string]string{{"id": "456", "username": "someone"}}, "meta": map[string]any{"result_count": 1, "next_token": "next-page"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()

	config := testConfig(api.URL)
	auth, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	transactionCookie, state := startLogin(t, auth)
	body := `{"code":"oauth-code","state":"` + state + `","redirectUri":"` + testRedirect + `"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/x/callback", strings.NewReader(body))
	req.AddCookie(transactionCookie)
	recorder := httptest.NewRecorder()
	auth.CallbackHandler(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("callback status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	if tokenCalls.Load() != 1 || followingCalls.Load() != 2 {
		t.Fatalf("token calls = %d, following calls = %d", tokenCalls.Load(), followingCalls.Load())
	}
	var sessionCookie *http.Cookie
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == sessionCookieName && cookie.Value != "" {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil || !sessionCookie.HttpOnly || sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("missing secure session cookie: %#v", recorder.Result().Cookies())
	}

	sessionReq := httptest.NewRequest(http.MethodGet, "/auth/x/session", nil)
	sessionReq.AddCookie(sessionCookie)
	sessionRecorder := httptest.NewRecorder()
	auth.SessionHandler(sessionRecorder, sessionReq)
	if sessionRecorder.Code != http.StatusOK || !strings.Contains(sessionRecorder.Body.String(), `"handle":"alice"`) {
		t.Fatalf("session response = %d %s", sessionRecorder.Code, sessionRecorder.Body.String())
	}

	identity, err := auth.Identity(sessionReq)
	if err != nil {
		t.Fatal(err)
	}
	if identity.UserID != "123" || identity.Handle != "alice" || identity.AccessToken != "user-token" {
		t.Fatalf("identity = %#v", identity)
	}

	restarted, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.Identity(sessionReq); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("session survived process-key rotation: %v", err)
	}

	logoutReq := httptest.NewRequest(http.MethodDelete, "/auth/x/session", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutRecorder := httptest.NewRecorder()
	auth.SessionHandler(logoutRecorder, logoutReq)
	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("logout status = %d", logoutRecorder.Code)
	}
	if _, err := auth.Identity(sessionReq); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("copied session remained valid after logout: %v", err)
	}
}

func TestCallbackTransactionCanOnlyBeConsumedOnce(t *testing.T) {
	var tokenCalls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenCalls.Add(1)
			writeTestJSON(w, http.StatusOK, map[string]any{"access_token": "user-token", "token_type": "bearer", "expires_in": 7200})
		case "/users/me":
			writeTestJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"id": "123", "username": "alice"}})
		case "/users/123/following":
			writeTestJSON(w, http.StatusOK, map[string]any{"data": []map[string]string{{"id": "999"}}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	auth, err := New(testConfig(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	cookie, state := startLogin(t, auth)
	body := `{"code":"code","state":"` + state + `","redirectUri":"` + testRedirect + `"}`
	for attempt := 0; attempt < 2; attempt++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/x/callback", strings.NewReader(body))
		req.AddCookie(cookie)
		recorder := httptest.NewRecorder()
		auth.CallbackHandler(recorder, req)
		if attempt == 0 && recorder.Code != http.StatusOK {
			t.Fatalf("first callback = %d %s", recorder.Code, recorder.Body.String())
		}
		if attempt == 1 && recorder.Code != http.StatusBadRequest {
			t.Fatalf("replayed callback = %d %s", recorder.Code, recorder.Body.String())
		}
	}
	if tokenCalls.Load() != 1 {
		t.Fatalf("token exchanges = %d", tokenCalls.Load())
	}
}

func TestCallbackRejectsUserWhoDoesNotFollowFlorin(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			writeTestJSON(w, http.StatusOK, map[string]any{"access_token": "user-token", "token_type": "bearer", "expires_in": 7200})
		case "/users/me":
			writeTestJSON(w, http.StatusOK, map[string]any{"data": map[string]string{"id": "123", "username": "alice"}})
		case "/users/123/following":
			writeTestJSON(w, http.StatusOK, map[string]any{"meta": map[string]any{"result_count": 0}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()
	auth, err := New(testConfig(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	cookie, state := startLogin(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/auth/x/callback", strings.NewReader(`{"code":"code","state":"`+state+`","redirectUri":"`+testRedirect+`"}`))
	req.AddCookie(cookie)
	recorder := httptest.NewRecorder()
	auth.CallbackHandler(recorder, req)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "@florinprotocol") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
	for _, responseCookie := range recorder.Result().Cookies() {
		if responseCookie.Name == sessionCookieName && responseCookie.Value != "" {
			t.Fatal("non-follower received a session")
		}
	}
}

func TestCallbackRejectsStateMismatchBeforeTokenExchange(t *testing.T) {
	var calls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "unexpected", http.StatusInternalServerError)
	}))
	defer api.Close()
	auth, err := New(testConfig(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	cookie, state := startLogin(t, auth)
	req := httptest.NewRequest(http.MethodPost, "/auth/x/callback", strings.NewReader(`{"code":"code","state":"wrong","redirectUri":"`+testRedirect+`"}`))
	req.AddCookie(cookie)
	recorder := httptest.NewRecorder()
	auth.CallbackHandler(recorder, req)
	if recorder.Code != http.StatusBadRequest || calls.Load() != 0 {
		t.Fatalf("status = %d, calls = %d", recorder.Code, calls.Load())
	}

	retry := httptest.NewRequest(http.MethodPost, "/auth/x/callback", strings.NewReader(`{"code":"code","state":"`+state+`","redirectUri":"`+testRedirect+`"}`))
	retry.AddCookie(cookie)
	retryRecorder := httptest.NewRecorder()
	auth.CallbackHandler(retryRecorder, retry)
	if retryRecorder.Code != http.StatusBadGateway || calls.Load() != 1 {
		t.Fatalf("valid retry status = %d, calls = %d", retryRecorder.Code, calls.Load())
	}
}

func TestVerifyFollowingFailsClosedOnRateLimit(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(w, http.StatusTooManyRequests, map[string]string{"title": "Too Many Requests"})
	}))
	defer api.Close()
	auth, err := New(testConfig(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	following, err := auth.VerifyFollowing(context.Background(), Identity{UserID: "123", AccessToken: "token", RequiredUserID: "999"})
	if err == nil || following {
		t.Fatalf("following = %v, err = %v", following, err)
	}
}

func TestMiddlewareAddsAValidIdentityToTheRequestContext(t *testing.T) {
	auth, err := New(testConfig("https://api.x.test"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := auth.seal(sessionPurpose, session{
		SessionID: "session-1", UserID: "123", Handle: "alice", AccessToken: "token", RequiredUserID: "999",
		ExpiresAt: auth.config.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatal(err)
	}
	var got Identity
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = IdentityFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodPost, "/query", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: value})
	recorder := httptest.NewRecorder()
	auth.Middleware(next).ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent || got.UserID != "123" || got.Handle != "alice" {
		t.Fatalf("status = %d, identity = %#v", recorder.Code, got)
	}
}

func TestSafeReturnToRejectsNetworkPathsAndControls(t *testing.T) {
	for _, value := range []string{"https://evil.example", "//evil.example", `/\\evil.example/path`, "/%5c%5cevil.example/path", "/claim\nheader"} {
		if got := safeReturnTo(value); got != "/" {
			t.Errorf("safeReturnTo(%q) = %q", value, got)
		}
	}
	if got := safeReturnTo("/claim?from=x#done"); got != "/claim?from=x#done" {
		t.Fatalf("valid return path = %q", got)
	}
}

func assertBearer(t *testing.T, r *http.Request, token string) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer "+token {
		t.Fatalf("authorization = %q", got)
	}
}

func writeTestJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
