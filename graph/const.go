package graph

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

var (
	FaucetAddr = ethCommon.HexToAddress("0x2916DD7fc1149f05eF2b462d1a3Ef4792929b550")
	ChainId = new(big.Int).SetInt64(46630)
)