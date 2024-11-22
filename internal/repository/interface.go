package repository

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines all database operations
type Repository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (GetUserByIDRow, error)
	GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
	UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (uuid.UUID, error)
	GetUserStats(ctx context.Context, id uuid.UUID) (GetUserStatsRow, error)
	GetUserRole(ctx context.Context, id uuid.UUID) (GetUserRoleRow, error)
	CheckUserIsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
	UpdateUserRole(ctx context.Context, arg UpdateUserRoleParams) ([]byte, error)
	ListUsersByRole(ctx context.Context, name string) ([]ListUsersByRoleRow, error)
}

// Ensure Queries implements Repository
var _ Repository = (*Queries)(nil)
