// auth/internal/usecase/validate.go
package usecase

import (
	"context"
	"fmt"

	"github.com/YaganovValera/analytics-system/common/ctxkeys"

	authpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/auth"
	commonpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/common"

	"github.com/YaganovValera/analytics-system/services/auth/internal/jwt"
	"github.com/YaganovValera/analytics-system/services/auth/internal/metrics"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type validateHandler struct {
	verifier jwt.Verifier
}

func NewValidateTokenHandler(verifier jwt.Verifier) ValidateTokenHandler {
	return &validateHandler{verifier}
}

func (h *validateHandler) Handle(ctx context.Context, token string) (*authpb.ValidateTokenResponse, error) {

	claims, err := h.verifier.Parse(token)
	if err != nil {
		metrics.ValidateTotal.WithLabelValues("invalid").Inc()
		return nil, fmt.Errorf("token invalid: %w", err)
	}
	metrics.ValidateTotal.WithLabelValues("ok").Inc()

	// Собираем metadata из контекста
	meta := &commonpb.RequestMetadata{
		TraceId:   getStringFromContext(ctx, ctxkeys.TraceIDKey),
		IpAddress: getStringFromContext(ctx, ctxkeys.IPAddressKey),
		UserAgent: getStringFromContext(ctx, ctxkeys.UserAgentKey),
	}

	return &authpb.ValidateTokenResponse{
		Valid:     true,
		Username:  claims.UserID,
		Roles:     claims.Roles,
		ExpiresAt: timestamppb.New(claims.ExpiresAt.Time),
		Metadata:  meta,
	}, nil
}

func getStringFromContext(ctx context.Context, key any) string {
	val, _ := ctx.Value(key).(string)
	return val
}
