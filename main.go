//go:generate go tool gqlgen generate

package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fluidity-money/faucet.florin.so/graph"
	"github.com/fluidity-money/faucet.florin.so/internal/xauth"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/lib/pq"
)

const (
	EnvTurnstileSecret    = "SPN_TURNSTILE_SECRET"
	EnvTurnstileHostnames = "SPN_TURNSTILE_HOSTNAMES"
	EnvGethAddr           = "SPN_ROBINHOOD_URL"
	EnvTimescaleURI       = "SPN_TIMESCALE"
	EnvXClientID          = "SPN_X_CLIENT_ID"
	EnvXClientSecret      = "SPN_X_CLIENT_SECRET"
	EnvXSessionSecret     = "SPN_X_SESSION_SECRET"
	EnvXRedirectURIs      = "SPN_X_REDIRECT_URIS"
)

const HttpUnixSocket = "/run/faucet.florin.so/http.sock"

func main() {
	turnstileSecret := os.Getenv(EnvTurnstileSecret)
	if turnstileSecret == "" {
		log.Fatal("turnstile unset")
	}
	turnstile, err := graph.NewTurnstileVerifier(
		turnstileSecret,
		splitCommaSeparated(os.Getenv(EnvTurnstileHostnames)),
	)
	if err != nil {
		log.Fatalf("configure Turnstile: %v", err)
	}

	xAuthenticator, err := xauth.New(xauth.Config{
		ClientID:         os.Getenv(EnvXClientID),
		ClientSecret:     os.Getenv(EnvXClientSecret),
		SessionSecret:    os.Getenv(EnvXSessionSecret),
		AllowedRedirects: splitCommaSeparated(os.Getenv(EnvXRedirectURIs)),
		SessionTTL:       2 * time.Hour,
		TransactionTTL:   10 * time.Minute,
	})
	if err != nil {
		log.Fatalf("configure X authentication: %v", err)
	}

	client, err := ethclient.Dial(os.Getenv(EnvGethAddr))
	if err != nil {
		log.Fatalf("ethclient dial: %v", err)
	}
	defer client.Close()
	db, err := sql.Open("postgres", os.Getenv(EnvTimescaleURI))
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{
			Turnstile: turnstile,
			Db:        db,
			Client:    client,
		},
	}))
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	mux := http.NewServeMux()
	mux.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	mux.HandleFunc("/auth/x/start", xAuthenticator.StartHandler)
	mux.HandleFunc("/auth/x/callback", xAuthenticator.CallbackHandler)
	mux.HandleFunc("/auth/x/session", xAuthenticator.SessionHandler)
	mux.Handle("/query", xAuthenticator.Middleware(srv))
	mux.Handle("/", xAuthenticator.Middleware(srv))
	_ = os.Remove(HttpUnixSocket)
	listener, err := net.Listen("unix", HttpUnixSocket)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	_ = os.Chmod(HttpUnixSocket, 0x0777)
	log.Printf("faucet listening on unix://%s", HttpUnixSocket)
	log.Fatal(http.Serve(listener, mux))
}

func splitCommaSeparated(value string) []string {
	var values []string
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}
