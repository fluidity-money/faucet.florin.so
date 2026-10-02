package graph

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

func canonicalRecipientAddress(value string) (common.Address, error) {
	if !common.IsHexAddress(value) {
		return common.Address{}, errors.New("bad recipient address")
	}
	return common.HexToAddress(value), nil
}
