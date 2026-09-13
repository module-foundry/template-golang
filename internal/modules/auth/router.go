package auth

import (
	"github.com/gofiber/fiber/v3"

	"template-golang/internal/config"
	"template-golang/pkg/httpx"
	"template-golang/pkg/jwtx"
)

// Deps lists the infrastructure the auth module needs. The module builds its
// own graph in Register; the application only passes shared dependencies.
type Deps struct {
	Config *config.Config
	Signer *jwtx.Signer
}

// Register is the composition root of the module: it creates the service and
// handler and mounts routes. Call it from internal/app/routes.go.
func Register(router fiber.Router, registry *httpx.Registry, deps Deps) {
	service := NewService(deps.Signer)
	handler := NewHandler(service, deps.Config)

	httpx.PostWithCtx[MiniAppTelegramRequest, MiniAppTelegramResponse](
		registry, router, "/auth/mini-apps/telegram", handler.Telegram, httpx.Tag("auth"),
	)
}
