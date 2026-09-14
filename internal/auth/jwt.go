package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"hzycoder.com/lion/internal/model"
)

const Issuer = "lion"

type Claims struct {
	UserID   int64      `json:"user_id"`
	Username string     `json:"username"`
	Role     model.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(secret []byte, expire time.Duration, userID int64, username string, role model.Role) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func ParseToken(secret []byte, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(Issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}
