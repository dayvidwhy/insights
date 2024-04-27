package auth

import (
	"testing"
)

func TestGenerateToken(t *testing.T) {
	_, err := generateToken(64)

	if err != nil {
		t.Fatalf("Token failed to generate.")
	}
}
