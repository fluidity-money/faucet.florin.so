package graph

import "testing"

func TestCanonicalRecipientAddressIgnoresInputCasing(t *testing.T) {
	lower, err := canonicalRecipientAddress("0xabcdefabcdefabcdefabcdefabcdefabcdefabcd")
	if err != nil {
		t.Fatal(err)
	}
	upper, err := canonicalRecipientAddress("0xABCDEFABCDEFABCDEFABCDEFABCDEFABCDEFABCD")
	if err != nil {
		t.Fatal(err)
	}
	if lower.Hex() != upper.Hex() {
		t.Fatalf("canonical forms differ: %q != %q", lower.Hex(), upper.Hex())
	}
}

func TestCanonicalRecipientAddressRejectsMalformedValue(t *testing.T) {
	if _, err := canonicalRecipientAddress("not-an-address"); err == nil {
		t.Fatal("expected malformed address to be rejected")
	}
}
