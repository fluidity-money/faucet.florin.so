package faucet

import (
	"strings"
	"math/big"

	ethAbi "github.com/ethereum/go-ethereum/accounts/abi"
)

var abi, _ = ethAbi.JSON(strings.NewReader(`[
  {
    "type": "function",
    "name": "receiveWeth",
    "inputs": [
      {
        "name": "_amt",
        "type": "uint256",
        "internalType": "uint256"
      }
    ],
    "outputs": [],
    "stateMutability": "nonpayable"
  }
]
`))

func MakeReceiveWethCd(amt *big.Int) []byte {
	cd, err := abi.Pack("receiveWeth", amt)
	if err != nil {
		panic(err)
	}
	return cd
}
