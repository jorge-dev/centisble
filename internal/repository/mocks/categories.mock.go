package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
)

type CategoryMock struct {
	categories map[string]repository.Category
}

func NewCategoryMock() *CategoryMock {
	return &CategoryMock{
		categories: make(map[string]repository.Category),
	}
}

func (m *CategoryMock) AddCategory(category repository.Category) {
	m.categories[category.ID.String()] = category
}

func (m *CategoryMock) CheckCategoryExists(ctx context.Context, arg repository.CheckCategoryExistsParams) (bool, error) {
	for _, cat := range m.categories {
		if cat.UserID == arg.UserID && cat.Name == arg.Name {
			return true, nil
		}
	}
	return false, nil
}

func (m *CategoryMock) CreateCategory(ctx context.Context, arg repository.CreateCategoryParams) (repository.Category, error) {
	now := time.Now()
	category := repository.Category{
		ID:        arg.ID,
		UserID:    arg.UserID,
		Name:      arg.Name,
		CreatedAt: now,
		UpdatedAt: &now,
	}
	m.categories[arg.ID.String()] = category
	return category, nil
}

func (m *CategoryMock) DeleteCategory(ctx context.Context, arg repository.DeleteCategoryParams) (int64, error) {
	key := arg.ID.String()
	if cat, exists := m.categories[key]; exists && cat.UserID == arg.UserID {
		now := time.Now()
		cat.DeletedAt = &now
		m.categories[key] = cat
		return 1, nil
	}
	return 0, nil
}

func (m *CategoryMock) GetCategoryByID(ctx context.Context, arg repository.GetCategoryByIDParams) (repository.Category, error) {
	if cat, exists := m.categories[arg.ID.String()]; exists && cat.UserID == arg.UserID {
		return cat, nil
	}
	return repository.Category{}, ErrRecordNotFound
}

func (m *CategoryMock) GetCategoryUsage(ctx context.Context, arg repository.GetCategoryUsageParams) (repository.GetCategoryUsageRow, error) {
	return repository.GetCategoryUsageRow{
		ID:            arg.ID,
		Name:          "Test Category",
		ExpenseCount:  5,
		BudgetCount:   2,
		TotalExpenses: 1000.00,
	}, nil
}

func (m *CategoryMock) GetMostUsedCategories(ctx context.Context, arg repository.GetMostUsedCategoriesParams) ([]repository.GetMostUsedCategoriesRow, error) {
	return []repository.GetMostUsedCategoriesRow{
		{
			Name:        "Test Category",
			UsageCount:  10,
			TotalAmount: 2000.00,
		},
	}, nil
}

func (m *CategoryMock) ListCategories(ctx context.Context, userID uuid.UUID) ([]repository.Category, error) {
	var result []repository.Category
	for _, cat := range m.categories {
		if cat.UserID == userID && cat.DeletedAt == nil {
			result = append(result, cat)
		}
	}
	return result, nil
}

func (m *CategoryMock) UpdateCategory(ctx context.Context, arg repository.UpdateCategoryParams) (repository.Category, error) {
	if cat, exists := m.categories[arg.ID.String()]; exists && cat.UserID == arg.UserID {
		now := time.Now()
		cat.Name = arg.Name
		cat.UpdatedAt = &now
		m.categories[arg.ID.String()] = cat
		return cat, nil
	}
	return repository.Category{}, ErrRecordNotFound
}
