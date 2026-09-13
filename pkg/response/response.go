// Package response writes the single JSON response shape used by handlers.
package response

import (
	"errors"

	"github.com/gofiber/fiber/v3"

	"template-golang/pkg/apperror"
)

// OK writes a 200 response with a JSON body.
func OK(c fiber.Ctx, v any) error {
	return c.Status(fiber.StatusOK).JSON(v)
}

// Created writes a 201 response with a JSON body.
func Created(c fiber.Ctx, v any) error {
	return c.Status(fiber.StatusCreated).JSON(v)
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
