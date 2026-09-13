package auth

import (
	"github.com/gofiber/fiber/v3"

	"template-golang/internal/config"
)

// Handler exposes the auth endpoints.
type Handler struct {
	service *Service
	cfg     *config.Config
}

// NewHandler builds the auth handler.
func NewHandler(service *Service, cfg *config.Config) *Handler {
	return &Handler{service: service, cfg: cfg}
}

// Telegram implements POST /auth/mini-apps/telegram.
func (h *Handler) Telegram(c fiber.Ctx, req MiniAppTelegramRequest) (MiniAppTelegramResponse, error) {
	_ = req
	resp, err := h.service.MiniAppTelegram(c.Context())
	if err != nil {
		return MiniAppTelegramResponse{}, err
	}

	maxAge, err := h.service.TokenTTL()
	if err != nil {
		return MiniAppTelegramResponse{}, err
	}
	c.Cookie(&fiber.Cookie{
		Name:     h.cfg.CookieName,
		Value:    resp.Token,
		Domain:   h.cfg.CookieDomain,
		MaxAge:   maxAge,
		Path:     "/",
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
	return resp, nil
}
