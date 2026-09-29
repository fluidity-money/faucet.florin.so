package graph

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

var (
	FaucetAddr = ethCommon.HexToAddress("0x8FF533385cA27d43c9Cc94D561a59384144f19F4")
	ChainId = new(big.Int).SetInt64(46630)
	EthToSend = new(big.Int).SetInt64(1)
)

func init() {
	EthToSend.Exp(EthToSend, new(big.Int).SetInt64(15), nil)
}
