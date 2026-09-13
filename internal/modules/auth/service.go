package auth

import (
	"context"
	"fmt"

	"uuid"

	"template-golang/pkg/apperror"
	"template-golang/pkg/jwtx"
)

// Service holds the business logic of the auth module.
type Service struct {
	signer *jwtx.Signer
}

// NewService builds the auth service.
func NewService(signer *jwtx.Signer) *Service {
	return &Service{signer: signer}
}

// MiniAppTelegram is a stub: it issues a token for a random user id.
// Replace the body with real Telegram initData validation in the future;
// the handler contract stays the same.
func (s *Service) MiniAppTelegram(ctx context.Context) (MiniAppTelegramResponse, error) {
	_ = ctx

	userID := uuid.New().String()
	token, err := s.signer.Sign(userID)
	if err != nil {
		return MiniAppTelegramResponse{}, fmt.Errorf("sign token: %w", err)
	}
	return MiniAppTelegramResponse{Token: token, UserID: userID}, nil
}

// TokenTTL returns the token lifetime from the signer.
func (s *Service) TokenTTL() (int, error) {
	ttl := s.signer.TTL().Seconds()
	if ttl <= 0 {
		return 0, apperror.Runtime(apperror.CodeConfigInvalid, fmt.Errorf("invalid jwt ttl"))
	}
	return int(ttl), nil
}
