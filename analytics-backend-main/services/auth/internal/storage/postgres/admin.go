// Файл: auth/internal/storage/postgres/admin.go
package postgres

import (
	"context"
	"fmt"

	commonpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/common"
)

// ListUsers возвращает пользователей постранично, с алфавитной пагинацией по username.
func (r *userRepo) ListUsers(ctx context.Context, page *commonpb.Pagination) ([]*User, string, error) {
	const defaultPageSize = 100
	pageSize := int32(defaultPageSize)
	if page != nil && page.PageSize > 0 {
		pageSize = page.PageSize
	}

	query := `
		SELECT id, username, password_hash, roles, created_at
		FROM users
		WHERE ($1::text IS NULL OR username > $1)
		ORDER BY username ASC
		LIMIT $2
	`

	var token any = nil
	if page != nil && page.PageToken != "" {
		token = page.PageToken
	}

	rows, err := r.db.Query(ctx, query, token, pageSize+1)
	if err != nil {
		return nil, "", fmt.Errorf("query users: %w", err)
	}
	defer rows.Close()

	var users []*User
	var nextToken string

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Roles, &u.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("scan user: %w", err)
		}
		users = append(users, &u)
	}

	if int32(len(users)) > pageSize {
		nextToken = users[pageSize].Username
		users = users[:pageSize]
	}

	return users, nextToken, rows.Err()
}

// UpdateUserRoles обновляет список ролей пользователя по ID.
func (r *userRepo) UpdateUserRoles(ctx context.Context, userID string, roles []string) error {
	const q = `UPDATE users SET roles = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, q, roles, userID)
	if err != nil {
		return fmt.Errorf("update roles: %w", err)
	}
	return nil
}

// GetUserByID возвращает одного пользователя по id.
func (r *userRepo) GetUserByID(ctx context.Context, userID string) (*User, error) {
	const q = `SELECT id, username, password_hash, roles, created_at FROM users WHERE id = $1`
	row := r.db.QueryRow(ctx, q, userID)

	var u User
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Roles, &u.CreatedAt); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}
