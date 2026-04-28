package auth_test

import (
	"bytes"
	"testing"

	"ralts-cms/pkg/auth"
)

func TestHashRefreshToken_Deterministic(t *testing.T) {
	a := auth.HashRefreshToken("raw", "pep")
	b := auth.HashRefreshToken("raw", "pep")
	if !bytes.Equal(a, b) {
		t.Fatalf("expected equal hashes")
	}
	if bytes.Equal(a, auth.HashRefreshToken("other", "pep")) {
		t.Fatalf("expected different hash for different raw")
	}
}

func TestGenerateAccessToken_RequiresPositiveTTL(t *testing.T) {
	_, err := auth.GenerateAccessToken(1, "ADMIN", "secret", 0)
	if err == nil {
		t.Fatalf("expected error for zero ttl")
	}
}

func TestGenerateRawRefreshToken_TwiceDiffers(t *testing.T) {
	a, err := auth.GenerateRawRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	b, err := auth.GenerateRawRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if a == b {
		t.Fatalf("expected two different tokens")
	}
}
