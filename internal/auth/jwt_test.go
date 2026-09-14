package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"hzycoder.com/lion/internal/auth"
	"hzycoder.com/lion/internal/model"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := []byte("test-secret")
	token, err := auth.GenerateToken(secret, time.Hour, 1, "demo", model.RoleMember)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	claims, err := auth.ParseToken(secret, token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}

	if claims.UserID != 1 || claims.Username != "demo" || claims.Role != model.RoleMember {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if claims.Issuer != auth.Issuer || claims.Subject != "1" {
		t.Fatalf("unexpected registered claims: %+v", claims.RegisteredClaims)
	}
}

func TestParseTokenRejectsWrongIssuer(t *testing.T) {
	secret := []byte("test-secret")
	claims := auth.Claims{
		UserID:   1,
		Username: "demo",
		Role:     model.RoleMember,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "evil-issuer",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := auth.ParseToken(secret, token); err == nil {
		t.Fatal("expected error for wrong issuer, got nil")
	}
}

func TestParseTokenRejectsExpired(t *testing.T) {
	secret := []byte("test-secret")
	token, err := auth.GenerateToken(secret, -time.Minute, 1, "demo", model.RoleMember)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	if _, err := auth.ParseToken(secret, token); err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestParseTokenRejectsAlgorithmConfusion(t *testing.T) {
	secret := []byte("test-secret")
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	claims := auth.Claims{
		UserID:   1,
		Username: "demo",
		Role:     model.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    auth.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := auth.ParseToken(secret, token); err == nil {
		t.Fatal("expected error for RS256 token, got nil")
	}
}
