package graph

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	turnstileAction         = "faucet"
	turnstileSiteverifyURL  = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	maxTurnstileTokenLength = 2048
	maxTurnstileResponse    = 64 << 10
)

var errTurnstileRejected = errors.New("Human verification failed. Please try again.")

type TurnstileVerifier struct {
	secret    string
	hostnames map[string]struct{}
	client    *http.Client
	endpoint  string
}

type turnstileResponse struct {
	Success    bool     `json:"success"`
	Action     string   `json:"action"`
	Hostname   string   `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
}

func NewTurnstileVerifier(secret string, hostnames []string) (*TurnstileVerifier, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, errors.New("Turnstile secret is required")
	}
	allowed := make(map[string]struct{}, len(hostnames))
	for _, hostname := range hostnames {
		if hostname = strings.ToLower(strings.TrimSpace(hostname)); hostname != "" {
			allowed[hostname] = struct{}{}
		}
	}
	if len(allowed) == 0 {
		return nil, errors.New("at least one Turnstile hostname is required")
	}
	return &TurnstileVerifier{
		secret:    secret,
		hostnames: allowed,
		client:    &http.Client{Timeout: 10 * time.Second},
		endpoint:  turnstileSiteverifyURL,
	}, nil
}

func (v *TurnstileVerifier) Verify(ctx context.Context, token string) error {
	if v == nil || token == "" || len(token) > maxTurnstileTokenLength {
		return errTurnstileRejected
	}

	form := url.Values{"secret": {v.secret}, "response": {token}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, v.endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		slog.Warn("Turnstile request creation failed", "err", err)
		return errTurnstileRejected
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := v.client.Do(request)
	if err != nil {
		slog.Warn("Turnstile siteverify request failed", "err", err)
		return errTurnstileRejected
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		slog.Warn("Turnstile siteverify returned a non-success status", "status", response.StatusCode)
		return errTurnstileRejected
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, maxTurnstileResponse+1))
	if err != nil || len(body) > maxTurnstileResponse {
		slog.Warn("Turnstile siteverify response could not be read", "err", err)
		return errTurnstileRejected
	}
	var result turnstileResponse
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Warn("Turnstile siteverify returned invalid JSON", "err", err)
		return errTurnstileRejected
	}
	_, hostnameAllowed := v.hostnames[strings.ToLower(result.Hostname)]
	if !result.Success || result.Action != turnstileAction || !hostnameAllowed {
		slog.Warn("Turnstile verification rejected", "action", result.Action, "hostname", result.Hostname, "error_codes", result.ErrorCodes)
		return errTurnstileRejected
	}
	return nil
}
