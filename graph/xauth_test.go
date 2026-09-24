package graph

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/fluidity-money/faucet.florin.so/internal/xauth"
)

type fakeFollowVerifier struct {
	follows bool
	err     error
	got     xauth.Identity
}

func (f *fakeFollowVerifier) VerifyFollowing(_ context.Context, identity xauth.Identity) (bool, error) {
	f.got = identity
	return f.follows, f.err
}

func TestAuthorizeXClaimRequiresVerifiedSession(t *testing.T) {
	resolver := &Resolver{XAuth: &fakeFollowVerifier{follows: true}}
	if err := resolver.authorizeXClaim(context.Background(), "alice"); err == nil || !strings.Contains(err.Error(), "Connect") {
		t.Fatalf("error = %v", err)
	}
}

func TestAuthorizeXClaimRejectsMismatchedHandle(t *testing.T) {
	verifier := &fakeFollowVerifier{follows: true}
	resolver := &Resolver{XAuth: verifier}
	ctx := xauth.ContextWithIdentity(context.Background(), xauth.Identity{UserID: "123", Handle: "alice", AccessToken: "token", RequiredUserID: "999"})
	if err := resolver.authorizeXClaim(ctx, "mallory"); err == nil || !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("error = %v", err)
	}
}

func TestAuthorizeXClaimChecksFollowStatusAtClaimTime(t *testing.T) {
	verifier := &fakeFollowVerifier{follows: false}
	resolver := &Resolver{XAuth: verifier}
	identity := xauth.Identity{UserID: "123", Handle: "Alice", AccessToken: "token", RequiredUserID: "999"}
	ctx := xauth.ContextWithIdentity(context.Background(), identity)
	if err := resolver.authorizeXClaim(ctx, "@alice"); err == nil || !strings.Contains(err.Error(), "@florinprotocol") {
		t.Fatalf("error = %v", err)
	}
	if verifier.got.UserID != identity.UserID {
		t.Fatalf("verified identity = %#v", verifier.got)
	}
}

func TestAuthorizeXClaimFailsClosedWhenXIsUnavailable(t *testing.T) {
	resolver := &Resolver{XAuth: &fakeFollowVerifier{err: errors.New("rate limited")}}
	ctx := xauth.ContextWithIdentity(context.Background(), xauth.Identity{UserID: "123", Handle: "alice", AccessToken: "token", RequiredUserID: "999"})
	if err := resolver.authorizeXClaim(ctx, "alice"); err == nil || !strings.Contains(err.Error(), "could not verify") {
		t.Fatalf("error = %v", err)
	}
}
