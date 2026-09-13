// Package response writes the single JSON response shape used by handlers.
package response

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
)

// Envelope is the single success shape: the payload is nested under "result".
// Wrapping keeps room for metadata (pagination, warnings) without breaking the
// payload contract, so error and success bodies stay distinguishable on the
// client.
type Envelope struct {
	Result any `json:"result"`
}

// OK writes a 200 response with the payload under "result".
func OK(c fiber.Ctx, v any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Result: v})
}

// Created writes a 201 response with the payload under "result".
func Created(c fiber.Ctx, v any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Result: v})
}

// NoContent writes a 204 response.
func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}

// Fail writes an error envelope directly, bypassing the global handler.
// Prefer returning errors from handlers so they are logged exactly once.
func Fail(c fiber.Ctx, err error) error {
	var appErr *apperror.Error
	if !errors.As(err, &appErr) {
		appErr = apperror.Runtime(apperror.CodeInternal, err)
	}
	body := apperror.Envelope{Error: apperror.Body{Type: appErr.Code, Message: appErr.PublicMessage()}}
	return c.Status(appErr.HTTPStatus()).JSON(body)
}
