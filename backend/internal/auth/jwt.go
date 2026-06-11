package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken creates a signed HS256 JWT for the given userID and role.
func GenerateToken(userID, role, secret string, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(expiry).Unix(),
		"iat":  time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("auth: sign token: %w", err)
	}

	return signed, nil
}

// ValidateToken parses and validates a JWT, returning the subject (userID) and role.
func ValidateToken(tokenStr, secret string) (userID, role string, err error) {
	token, parseErr := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("auth: unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if parseErr != nil {
		return "", "", fmt.Errorf("auth: parse token: %w", parseErr)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", errors.New("auth: invalid token claims")
	}

	sub, subErr := claims.GetSubject()
	if subErr != nil || sub == "" {
		return "", "", errors.New("auth: missing subject claim")
	}

	r, _ := claims["role"].(string)
	if r == "" {
		r = "user"
	}

	return sub, r, nil
}
