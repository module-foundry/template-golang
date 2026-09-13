// Package middleware contains the few custom Fiber middlewares of the project.
package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"
)

type ctxKey int

const (
	userIDKey ctxKey = iota
)

// WithUserID stores the authenticated user id in a context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromCtx reads the authenticated user id.
func UserIDFromCtx(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok && userID != ""
}

// UserIDFromFiber is a convenience for handlers with a Fiber context.
func UserIDFromFiber(c fiber.Ctx) (string, bool) {
	return UserIDFromCtx(c.Context())
}
