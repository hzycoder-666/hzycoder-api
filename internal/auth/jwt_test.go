package auth_test

import (
	"testing"
	"time"

	"hzycoder.com/go-gin-template/internal/auth"
	"hzycoder.com/go-gin-template/internal/model"
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
}
