package graph

import (
	"database/sql"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Resolver struct{
	Db *sql.DB
	TurnstileSecret string
	Client *ethclient.Client
}
