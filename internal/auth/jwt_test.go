package auth

import (
	"os"
	"testing"
)

func TestGenerateParse_WithEnvAndAudienceIssuer(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_ISS", "ts")
	os.Setenv("JWT_AUD", "ts-clients")
	os.Setenv("JWT_TTL_MINUTES", "1")

	tok, err := GenerateJWT(7, "amir", "admin")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	claims, err := ParseJWT(tok)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims["username"] != "amir" || claims["role"] != "admin" {
		t.Fatalf("claims mismatch: %v", claims)
	}
}
