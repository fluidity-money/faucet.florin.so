package graph

import (
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
)

var (
	FaucetAddr = ethCommon.HexToAddress("0xC82B04cba01D5D919fc292a983481e9c85e51579")
	ChainId = new(big.Int).SetInt64(46630)
)