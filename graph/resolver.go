package graph

import (
	"context"
	"database/sql"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fluidity-money/faucet.florin.so/internal/xauth"
)

type followVerifier interface {
	VerifyFollowing(context.Context, xauth.Identity) (bool, error)
}

type Resolver struct {
	Db              *sql.DB
	TurnstileSecret string
	Client          *ethclient.Client
	XAuth           followVerifier
}
