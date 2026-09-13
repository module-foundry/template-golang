package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jsonx"
	"template-golang/pkg/jwtx"
)

func testApp(signer *jwtx.Signer) *fiber.App {
	app := fiber.New(fiber.Config{
		JSONEncoder:  jsonx.FiberEncoder,
		JSONDecoder:  jsonx.FiberDecoder,
		ErrorHandler: apperror.Handler(nil),
	})
	protected := app.Group("/", Auth(signer))
	protected.Get("/me", func(c fiber.Ctx) error {
		userID, ok := UserIDFromFiber(c)
		if !ok {
			return apperror.API(apperror.CodeUnauthorized, "no user")
		}
		return c.JSON(map[string]string{"user_id": userID})
	})
	return app
}

func TestAuthFromCookie(t *testing.T) {
	signer := jwtx.New("secret", time.Hour, "access_token")
	token, err := signer.Sign("user-1")
	if err != nil {
		t.Fatal(err)
	}
	app := testApp(signer)

	req := httptest.NewRequest(fiber.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestAuthFromHeader(t *testing.T) {
	signer := jwtx.New("secret", time.Hour, "access_token")
	token, err := signer.Sign("user-2")
	if err != nil {
		t.Fatal(err)
	}
	app := testApp(signer)

	req := httptest.NewRequest(fiber.MethodGet, "/me", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}
}

func TestAuthMissingToken(t *testing.T) {
	signer := jwtx.New("secret", time.Hour, "access_token")
	app := testApp(signer)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/me", nil))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 401 {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestAuthBadToken(t *testing.T) {
	signer := jwtx.New("secret", time.Hour, "access_token")
	app := testApp(signer)

	req := httptest.NewRequest(fiber.MethodGet, "/me", nil)
	req.Header.Set(fiber.HeaderAuthorization, "Bearer garbage")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != 401 {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}
