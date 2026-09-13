package apperror

import (
	"errors"
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v3"
)

// Envelope is the single error response shape returned to clients.
type Envelope struct {
	Error Body `json:"error"`
}

// Body carries the stable code and a human-readable message.
type Body struct {
	Type    Code   `json:"type"`
	Message string `json:"message"`
}

// Handler is the Fiber ErrorHandler. It converts every error into the stable
// envelope and logs it exactly once (API -> warn, Runtime -> error).
func Handler(log *slog.Logger) fiber.ErrorHandler {
	if log == nil {
		log = slog.Default()
	}
	return func(c fiber.Ctx, err error) error {
		var appErr *Error
		if !errors.As(err, &appErr) {
			appErr = Runtime(CodeInternal, err)
		}

		requestID, _ := c.Locals("request_id").(string)
		userID, _ := c.Locals("user_id").(string)

		attrs := []any{
			slog.String("type", string(appErr.Code)),
			slog.String("message", appErr.Message),
			slog.String("trace", appErr.Trace()),
			slog.String("kind", string(appErr.Kind())),
			slog.Int("http_status", appErr.HTTPStatus()),
			slog.String("request_id", requestID),
			slog.String("user_id", userID),
		}
		if appErr.cause != nil {
			attrs = append(attrs, slog.String("error", appErr.cause.Error()))
		}
		if appErr.Kind() == KindRuntime {
			if StackEnabled() {
				attrs = append(attrs, slog.String("stack", string(debug.Stack())))
			}
			log.Error("request failed", attrs...)
		} else {
			log.Warn("request failed", attrs...)
		}

		body := Envelope{Error: Body{Type: appErr.Code, Message: appErr.PublicMessage()}}
		return c.Status(appErr.HTTPStatus()).JSON(body)
	}
}
