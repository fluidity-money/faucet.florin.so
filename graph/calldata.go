package graph

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// faucetCalldata encodes the recipient and 100 tokens with 18 decimals.
func faucetCalldata(recipient common.Address) []byte {
	amount := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	amount.Mul(amount, big.NewInt(100))

	data := make([]byte, 4+32+32)
	copy(data[:4], []byte{0xd0, 0x67, 0x9d, 0x34})
	copy(data[4+12:4+32], recipient[:])
	amount.FillBytes(data[4+32:])
	return data
}
