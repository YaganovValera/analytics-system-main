// Файл: api-gateway/internal/handler/admin_handlers.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/YaganovValera/analytics-system/services/api-gateway/internal/response"

	authpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/auth"
	commonpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/common"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	pageSize := parseInt32Query(r, "page_size", 50)
	pageToken := r.URL.Query().Get("page_token")

	resp, err := h.Auth.ListUsers(r.Context(), &authpb.ListUsersRequest{
		Pagination: &commonpb.Pagination{
			PageSize:  pageSize,
			PageToken: pageToken,
		},
	})
	if err != nil {
		response.InternalError(w, "failed to list users"+err.Error())
		return
	}
	response.JSON(w, resp)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		response.BadRequest(w, "missing user id")
		return
	}
	resp, err := h.Auth.GetUser(r.Context(), &authpb.GetUserRequest{UserId: userID})
	if err != nil {
		response.InternalError(w, "get user failed"+err.Error())
		return
	}
	response.JSON(w, resp.User)
}

func (h *Handler) UpdateUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		response.BadRequest(w, "missing user id")
		return
	}
	var req struct {
		Roles []string `json:"roles"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Roles) == 0 {
		response.BadRequest(w, "invalid roles payload")
		return
	}
	_, err := h.Auth.UpdateUserRoles(r.Context(), &authpb.UpdateUserRolesRequest{
		UserId: userID,
		Roles:  req.Roles,
	})
	if err != nil {
		response.InternalError(w, "update roles failed"+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) AdminRevokeToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		response.BadRequest(w, "missing token")
		return
	}
	_, err := h.Auth.RevokeToken(r.Context(), &authpb.RevokeTokenRequest{
		Token: req.Token,
		Type:  authpb.TokenType_REFRESH,
	})
	if err != nil {
		response.InternalError(w, "revoke failed"+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
