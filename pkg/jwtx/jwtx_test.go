package jwtx

import (
	"errors"
	"testing"
	"time"

	"template-golang/pkg/apperror"
)

func TestSignAndParse(t *testing.T) {
	signer := New("secret", time.Hour, "access_token")
	token, err := signer.Sign("user-1")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	claims, err := signer.Parse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("user id = %q", claims.UserID)
	}
	if claims.SignAt == 0 {
		t.Fatal("sign_at must be set")
	}
}

func TestParseExpiredToken(t *testing.T) {
	signer := New("secret", -time.Minute, "access_token")
	token, err := signer.Sign("user-1")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = signer.Parse(token)
	if !hasCode(err, apperror.CodeTokenExpired) {
		t.Fatalf("want TOKEN_EXPIRED, got %v", err)
	}
}

func TestParseWrongSecret(t *testing.T) {
	issuer := New("secret-a", time.Hour, "access_token")
	verifier := New("secret-b", time.Hour, "access_token")
	token, err := issuer.Sign("user-1")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	_, err = verifier.Parse(token)
	if !hasCode(err, apperror.CodeUnauthorized) {
		t.Fatalf("want UNAUTHORIZED, got %v", err)
	}
}

func TestExtractPrefersCookie(t *testing.T) {
	if got := Extract("cookie-token", "Bearer header-token"); got != "cookie-token" {
		t.Fatalf("got %q", got)
	}
	if got := Extract("", "Bearer header-token"); got != "header-token" {
		t.Fatalf("got %q", got)
	}
	if got := Extract("", ""); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := Extract("", "Basic abc"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func BenchmarkSignParse(b *testing.B) {
	signer := New("secret", time.Hour, "access_token")
	token, err := signer.Sign("user-1")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := signer.Parse(token); err != nil {
			b.Fatal(err)
		}
	}
}

func hasCode(err error, code apperror.Code) bool {
	var appErr *apperror.Error
	return errors.As(err, &appErr) && appErr.Code == code
}
