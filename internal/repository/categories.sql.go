// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: categories.sql

package repository

import (
	"context"

	"github.com/google/uuid"
)

const checkCategoryExists = `-- name: CheckCategoryExists :one
SELECT EXISTS(
    SELECT 1 FROM categories
    WHERE user_id = $1 AND name = $2 AND deleted_at IS NULL
)
`

type CheckCategoryExistsParams struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

func (q *Queries) CheckCategoryExists(ctx context.Context, arg CheckCategoryExistsParams) (bool, error) {
	row := q.db.QueryRow(ctx, checkCategoryExists, arg.UserID, arg.Name)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const createCategory = `-- name: CreateCategory :one
INSERT INTO categories (id, user_id, name, created_at, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, user_id, name, created_at, updated_at, deleted_at
`

type CreateCategoryParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
}

func (q *Queries) CreateCategory(ctx context.Context, arg CreateCategoryParams) (Category, error) {
	row := q.db.QueryRow(ctx, createCategory, arg.ID, arg.UserID, arg.Name)
	var i Category
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const deleteCategory = `-- name: DeleteCategory :execrows
UPDATE categories 
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type DeleteCategoryParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) DeleteCategory(ctx context.Context, arg DeleteCategoryParams) (int64, error) {
	result, err := q.db.Exec(ctx, deleteCategory, arg.ID, arg.UserID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const getCategoryByID = `-- name: GetCategoryByID :one
SELECT id, user_id, name, created_at, updated_at, deleted_at FROM categories
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
`

type GetCategoryByIDParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) GetCategoryByID(ctx context.Context, arg GetCategoryByIDParams) (Category, error) {
	row := q.db.QueryRow(ctx, getCategoryByID, arg.ID, arg.UserID)
	var i Category
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}

const getCategoryUsage = `-- name: GetCategoryUsage :one
SELECT 
    c.id,
    c.name,
    COUNT(DISTINCT e.id) as expense_count,
    COUNT(DISTINCT b.id) as budget_count,
    COALESCE(SUM(e.amount), 0) as total_expenses
FROM categories c
LEFT JOIN expenses e ON 
     e.category_id = c.id
    AND e.user_id = c.user_id 
    AND e.deleted_at IS NULL
LEFT JOIN budgets b ON 
     b.category_id = c.id 
    AND b.user_id = c.user_id 
    AND b.deleted_at IS NULL
WHERE c.id = $1 AND c.user_id = $2 AND c.deleted_at IS NULL
GROUP BY c.id, c.name
`

type GetCategoryUsageParams struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
}

type GetCategoryUsageRow struct {
	ID            uuid.UUID   `json:"id"`
	Name          string      `json:"name"`
	ExpenseCount  int64       `json:"expense_count"`
	BudgetCount   int64       `json:"budget_count"`
	TotalExpenses interface{} `json:"total_expenses"`
}

func (q *Queries) GetCategoryUsage(ctx context.Context, arg GetCategoryUsageParams) (GetCategoryUsageRow, error) {
	row := q.db.QueryRow(ctx, getCategoryUsage, arg.ID, arg.UserID)
	var i GetCategoryUsageRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.ExpenseCount,
		&i.BudgetCount,
		&i.TotalExpenses,
	)
	return i, err
}

const getMostUsedCategories = `-- name: GetMostUsedCategories :many
SELECT 
    c.name,
    COUNT(e.id) as usage_count,
    COALESCE(SUM(e.amount), 0) as total_amount
FROM categories c
LEFT JOIN expenses e ON 
    e.category_id = c.id 
    AND e.user_id = c.user_id 
    AND e.deleted_at IS NULL
WHERE c.user_id = $1::UUID
    AND c.deleted_at IS NULL
GROUP BY c.name
ORDER BY usage_count DESC
LIMIT $2::int
`

type GetMostUsedCategoriesParams struct {
	ID    uuid.UUID `json:"id"`
	Limit int32     `json:"limit"`
}

type GetMostUsedCategoriesRow struct {
	Name        string      `json:"name"`
	UsageCount  int64       `json:"usage_count"`
	TotalAmount interface{} `json:"total_amount"`
}

func (q *Queries) GetMostUsedCategories(ctx context.Context, arg GetMostUsedCategoriesParams) ([]GetMostUsedCategoriesRow, error) {
	rows, err := q.db.Query(ctx, getMostUsedCategories, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMostUsedCategoriesRow
	for rows.Next() {
		var i GetMostUsedCategoriesRow
		if err := rows.Scan(&i.Name, &i.UsageCount, &i.TotalAmount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listCategories = `-- name: ListCategories :many
SELECT id, user_id, name, created_at, updated_at, deleted_at FROM categories
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name ASC
`

func (q *Queries) ListCategories(ctx context.Context, userID uuid.UUID) ([]Category, error) {
	rows, err := q.db.Query(ctx, listCategories, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var i Category
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Name,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.DeletedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateCategory = `-- name: UpdateCategory :one
UPDATE categories 
SET 
    name = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $3 AND deleted_at IS NULL
RETURNING id, user_id, name, created_at, updated_at, deleted_at
`

type UpdateCategoryParams struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	UserID uuid.UUID `json:"user_id"`
}

func (q *Queries) UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error) {
	row := q.db.QueryRow(ctx, updateCategory, arg.ID, arg.Name, arg.UserID)
	var i Category
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.Name,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
	)
	return i, err
}
