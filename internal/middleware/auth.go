package middleware

import (
	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jwtx"
)

// Auth validates the JWT from the cookie first, then the Authorization header.
// It is mounted only on the protected group: /auth/* stays public by design.
func Auth(signer *jwtx.Signer) fiber.Handler {
	return func(c fiber.Ctx) error {
		raw := jwtx.Extract(c.Cookies(signer.CookieName()), c.Get(fiber.HeaderAuthorization))
		if raw == "" {
			return apperror.API(apperror.CodeUnauthorized, "authentication required")
		}
		claims, err := signer.Parse(raw)
		if err != nil {
			return err
		}

		ctx := WithUserID(c.Context(), claims.UserID)
		c.SetContext(ctx)
		c.Locals("user_id", claims.UserID)
		return c.Next()
	}
}
