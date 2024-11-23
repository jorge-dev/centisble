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

	// Budget operations
	CreateBudget(ctx context.Context, arg CreateBudgetParams) (Budget, error)
	DeleteBudget(ctx context.Context, arg DeleteBudgetParams) (int64, error)
	GetActiveBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	GetBudgetByID(ctx context.Context, arg GetBudgetByIDParams) (Budget, error)
	GetBudgetUsage(ctx context.Context, arg GetBudgetUsageParams) (GetBudgetUsageRow, error)
	GetBudgetsByCategory(ctx context.Context, arg GetBudgetsByCategoryParams) ([]Budget, error)
	GetBudgetsNearLimit(ctx context.Context, arg GetBudgetsNearLimitParams) ([]GetBudgetsNearLimitRow, error)
	GetOneTimeBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	GetRecurringBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	ListBudgets(ctx context.Context, userID uuid.UUID) ([]Budget, error)
	UpdateBudget(ctx context.Context, arg UpdateBudgetParams) (Budget, error)

	// Category operations
	CheckCategoryExists(ctx context.Context, arg CheckCategoryExistsParams) (bool, error)
	CreateCategory(ctx context.Context, arg CreateCategoryParams) (Category, error)
	DeleteCategory(ctx context.Context, arg DeleteCategoryParams) (int64, error)
	GetCategoryByID(ctx context.Context, arg GetCategoryByIDParams) (Category, error)
	GetCategoryUsage(ctx context.Context, arg GetCategoryUsageParams) (GetCategoryUsageRow, error)
	GetMostUsedCategories(ctx context.Context, arg GetMostUsedCategoriesParams) ([]GetMostUsedCategoriesRow, error)
	ListCategories(ctx context.Context, userID uuid.UUID) ([]Category, error)
	UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error)
}

// Ensure Queries implements Repository
var _ Repository = (*Queries)(nil)
