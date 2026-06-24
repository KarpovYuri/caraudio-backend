package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

const RoleAdmin = "admin"

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwtlib.RegisteredClaims
}

func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwtlib.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwtlib.ErrTokenInvalidClaims
	}

	return claims, nil
}

func ValidateAdmin(tokenString, secret string) error {
	if tokenString == "" {
		return ErrUnauthorized
	}

	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		return ErrUnauthorized
	}
	if claims.Role != RoleAdmin {
		return ErrForbidden
	}
	return nil
}

// Unused but keeps parity with auth token generation for tests.
func GenerateToken(userID, role, secret string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
		},
	}
	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
