// auth/internal/storage/postgres/interface.go
package postgres

import (
	"context"
	"time"

	commonpb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/common"
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
	Roles        []string
	CreatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	JTI       string
	Token     string
	IssuedAt  time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) error
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	ListUsers(ctx context.Context, page *commonpb.Pagination) ([]*User, string, error)
	UpdateUserRoles(ctx context.Context, userID string, roles []string) error
	GetUserByID(ctx context.Context, userID string) (*User, error)
}

type TokenRepository interface {
	Store(ctx context.Context, token *RefreshToken) error
	FindByJTI(ctx context.Context, jti string) (*RefreshToken, error)
	RevokeByJTI(ctx context.Context, jti string) error
}
