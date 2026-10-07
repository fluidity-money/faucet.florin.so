package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTurnstileVerifierAcceptsExpectedActionAndHostname(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.Form.Get("secret"); got != "secret" {
			t.Errorf("secret = %q", got)
		}
		if got := r.Form.Get("response"); got != "valid-token" {
			t.Errorf("response = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"success":  true,
			"action":   turnstileAction,
			"hostname": "faucet.florin.so",
		})
	}))
	defer server.Close()

	verifier, err := NewTurnstileVerifier("secret", []string{"faucet.florin.so"})
	if err != nil {
		t.Fatal(err)
	}
	verifier.endpoint = server.URL
	if err := verifier.Verify(context.Background(), "valid-token"); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
}

func TestTurnstileVerifierRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]any
	}{
		{name: "unsuccessful", response: map[string]any{"success": false}},
		{name: "wrong action", response: map[string]any{"success": true, "action": "other", "hostname": "faucet.florin.so"}},
		{name: "wrong hostname", response: map[string]any{"success": true, "action": turnstileAction, "hostname": "attacker.example"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_ = json.NewEncoder(w).Encode(test.response)
			}))
			defer server.Close()
			verifier, err := NewTurnstileVerifier("secret", []string{"faucet.florin.so"})
			if err != nil {
				t.Fatal(err)
			}
			verifier.endpoint = server.URL
			if err := verifier.Verify(context.Background(), "token"); err == nil {
				t.Fatal("Verify() unexpectedly accepted response")
			}
		})
	}
}

func TestTurnstileVerifierRejectsMissingOrOversizedTokenWithoutCallingCloudflare(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	verifier, err := NewTurnstileVerifier("secret", []string{"faucet.florin.so"})
	if err != nil {
		t.Fatal(err)
	}
	verifier.endpoint = server.URL
	for _, token := range []string{"", string(make([]byte, maxTurnstileTokenLength+1))} {
		if err := verifier.Verify(context.Background(), token); err == nil {
			t.Fatal("Verify() unexpectedly accepted invalid token")
		}
	}
	if calls != 0 {
		t.Fatalf("Cloudflare called %d times", calls)
	}
}

func TestNewTurnstileVerifierRequiresSecretAndHostnames(t *testing.T) {
	if _, err := NewTurnstileVerifier("", []string{"faucet.florin.so"}); err == nil {
		t.Fatal("missing secret was accepted")
	}
	if _, err := NewTurnstileVerifier("secret", nil); err == nil {
		t.Fatal("missing hostname allowlist was accepted")
	}
}
