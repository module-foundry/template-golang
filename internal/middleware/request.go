package middleware

import (
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/logger"
	"uuid"
)

// RequestID assigns a request id, exposes it in the response and in logs.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("X-Request-ID", requestID)
		c.Locals("request_id", requestID)
		return c.Next()
	}
}

// RequestLogger attaches a request-scoped logger to the context and logs
// a single access line per request.
func RequestLogger(log *slog.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		requestID, _ := c.Locals("request_id").(string)
		requestLog := log.With(slog.String("request_id", requestID))
		c.SetContext(logger.WithContext(c.Context(), requestLog))
		c.Locals("logger", requestLog)

		err := c.Next()

		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", c.Response().StatusCode()),
			slog.Int64("latency_ms", time.Since(start).Milliseconds()),
		}
		if userID, ok := UserIDFromFiber(c); ok {
			attrs = append(attrs, slog.String("user_id", userID))
		}

		switch {
		case err != nil || c.Response().StatusCode() >= 500:
			log.Error("access", attrs...)
		case c.Response().StatusCode() >= 400:
			log.Warn("access", attrs...)
		default:
			log.Info("access", attrs...)
		}
		return err
	}
}
