package jwtx

import (
	"testing"
	"time"

	"template-golang/pkg/apperror"
)

func TestSignerAccessors(t *testing.T) {
	signer := New("secret", 90*time.Minute, "token_cookie")
	if signer.TTL() != 90*time.Minute {
		t.Fatalf("ttl = %s", signer.TTL())
	}
	if signer.CookieName() != "token_cookie" {
		t.Fatalf("cookie = %q", signer.CookieName())
	}
}

func TestParseMalformedToken(t *testing.T) {
	signer := New("secret", time.Hour, "access_token")
	_, err := signer.Parse("not-a-jwt")
	if !hasCode(err, apperror.CodeUnauthorized) {
		t.Fatalf("want UNAUTHORIZED, got %v", err)
	}
}

func TestExtractBearerWithSpaces(t *testing.T) {
	if got := Extract("", "Bearer   spaced"); got != "spaced" {
		t.Fatalf("got %q", got)
	}
}
