// Package jwtx signs and parses the project JWT and extracts tokens from
// requests. Claims carry user_id and sign_at as required by the API contract.
package jwtx

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"template-golang/pkg/apperror"
)

// Claims is the token payload. sign_at is the issuance timestamp.
type Claims struct {
	UserID string `json:"user_id"`
	SignAt int64  `json:"sign_at"`
	jwt.RegisteredClaims
}

// Signer issues and parses tokens with a single secret.
type Signer struct {
	secret     []byte
	ttl        time.Duration
	cookieName string
}

// New builds a Signer.
func New(secret string, ttl time.Duration, cookieName string) *Signer {
	return &Signer{secret: []byte(secret), ttl: ttl, cookieName: cookieName}
}

// TTL returns the configured token lifetime.
func (s *Signer) TTL() time.Duration { return s.ttl }

// CookieName returns the cookie used to carry the token.
func (s *Signer) CookieName() string { return s.cookieName }

// Sign issues a token for the user.
func (s *Signer) Sign(userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		SignAt: now.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", apperror.Runtime(apperror.CodeInternal, err)
	}
	return signed, nil
}

// Parse validates a token and returns its claims. Expired tokens map to
// TOKEN_EXPIRED, everything else to UNAUTHORIZED.
func (s *Signer) Parse(raw string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperror.API(apperror.CodeTokenExpired, "token expired")
		}
		return nil, apperror.API(apperror.CodeUnauthorized, "invalid token")
	}
	return claims, nil
}

// Extract resolves a token from the cookie first, then the Authorization header.
func Extract(cookieValue, authorization string) string {
	if cookieValue != "" {
		return cookieValue
	}
	const prefix = "Bearer "
	if strings.HasPrefix(authorization, prefix) {
		return strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	}
	return ""
}
