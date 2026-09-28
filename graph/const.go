package graph

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

var (
	FaucetAddr = ethCommon.HexToAddress("0x1F3A25eC28c8E07c89591e30ed4fa19fFdE46C6D")
	ChainId = new(big.Int).SetInt64(46630)
)