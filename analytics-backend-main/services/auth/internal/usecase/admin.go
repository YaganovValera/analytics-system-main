// Файл: auth/internal/usecase/admin.go
package usecase

import (
	"context"
	"fmt"

	authpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/auth"
	"github.com/YaganovValera/analytics-system/services/auth/internal/jwt"
	"github.com/YaganovValera/analytics-system/services/auth/internal/storage/postgres"
)

type AdminHandler interface {
	ListUsers(ctx context.Context, req *authpb.ListUsersRequest) (*authpb.ListUsersResponse, error)
	UpdateUserRoles(ctx context.Context, req *authpb.UpdateUserRolesRequest) (*authpb.UpdateUserRolesResponse, error)
	GetUser(ctx context.Context, req *authpb.GetUserRequest) (*authpb.GetUserResponse, error)
}

type adminHandler struct {
	repo postgres.UserRepository
}

func NewAdminHandler(repo postgres.UserRepository) AdminHandler {
	return &adminHandler{repo: repo}
}

func (h *adminHandler) ListUsers(ctx context.Context, req *authpb.ListUsersRequest) (*authpb.ListUsersResponse, error) {
	users, next, err := h.repo.ListUsers(ctx, req.GetPagination())
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	resp := &authpb.ListUsersResponse{
		NextPageToken: next,
	}
	for _, u := range users {
		resp.Users = append(resp.Users, &authpb.UserInfo{
			Id:        u.ID,
			Username:  u.Username,
			Roles:     u.Roles,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return resp, nil
}

func (h *adminHandler) UpdateUserRoles(ctx context.Context, req *authpb.UpdateUserRolesRequest) (*authpb.UpdateUserRolesResponse, error) {
	if req == nil || req.UserId == "" || len(req.Roles) == 0 {
		return nil, fmt.Errorf("invalid request")
	}
	roles, err := jwt.NormalizeRoles(req.Roles)
	if err != nil {
		return nil, fmt.Errorf("invalid roles: %w", err)
	}
	if err := h.repo.UpdateUserRoles(ctx, req.UserId, roles); err != nil {
		return nil, fmt.Errorf("update roles: %w", err)
	}
	return &authpb.UpdateUserRolesResponse{Success: true}, nil
}

func (h *adminHandler) GetUser(ctx context.Context, req *authpb.GetUserRequest) (*authpb.GetUserResponse, error) {
	if req == nil || req.UserId == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	user, err := h.repo.GetUserByID(ctx, req.UserId)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &authpb.GetUserResponse{
		User: &authpb.UserInfo{
			Id:        user.ID,
			Username:  user.Username,
			Roles:     user.Roles,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}
