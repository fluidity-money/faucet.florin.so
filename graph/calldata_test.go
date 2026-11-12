package graph

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestFaucetCalldata(t *testing.T) {
	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
	got := hex.EncodeToString(faucetCalldata(recipient))
	// Selector, ABI-padded recipient, and 100 SPY (100000000000000000000).
	want := "d0679d34" +
		"0000000000000000000000001234567890123456789012345678901234567890" +
		"0000000000000000000000000000000000000000000000056bc75e2d63100000"
	if got != want {
		t.Fatalf("calldata = %s, want %s", got, want)
	}
}
