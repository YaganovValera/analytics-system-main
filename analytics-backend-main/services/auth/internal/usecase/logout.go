// auth/internal/usecase/logout.go
package usecase

import (
	"context"

	"github.com/YaganovValera/analytics-system/services/auth/internal/storage/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type logoutHandler struct {
	tokens postgres.TokenRepository
}

func NewLogoutHandler(tokens postgres.TokenRepository) LogoutHandler {
	return &logoutHandler{tokens}
}

func (h *logoutHandler) Handle(ctx context.Context, jti string) error {

	token, err := h.tokens.FindByJTI(ctx, jti)
	if err != nil {
		return status.Errorf(codes.NotFound, "refresh token not found: %v", err)
	}
	if token.RevokedAt != nil {
		return status.Error(codes.AlreadyExists, "token already revoked")
	}
	return h.tokens.RevokeByJTI(ctx, jti)
}
