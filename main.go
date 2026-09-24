//go:generate go tool gqlgen generate

package main

import (
	"net"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/fluidity-money/faucet.florin.so/graph"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	EnvTurnstileSecret = "SPN_TURNSTILE_SECRET"
	EnvGethAddr        = "SPN_ROBINHOOD_URL"
	EnvTimescaleUri = "SPN_TIMESCALE"
)

const HttpUnixSocket = "/var/run/faucet.florin.so/http.sock"

func main() {
	turnstileSecret := os.Getenv(EnvTurnstileSecret)
	if turnstileSecret == "" {
		log.Fatalf("turnstile unset")
	}
	client, err := ethclient.Dial(os.Getenv(EnvGethAddr))
	if err != nil {
		log.Fatalf("ethclient dial: %v", err)
	}
	defer client.Close()
	db, err := sql.Open("postgres", os.Getenv(EnvTimescaleUri))
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{
			TurnstileSecret: turnstileSecret,
			Db:              db,
			Client:          client,
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
	http.Handle("/playground", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/", srv)
	_ = os.Remove(HttpUnixSocket)
	l, err := net.Listen("unix", HttpUnixSocket)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	panic(http.Serve(l, nil))
}
