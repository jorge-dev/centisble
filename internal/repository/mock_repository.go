package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrRecordNotFound = errors.New("record not found")

// MockRepository implements Repository interface for testing
type MockRepository struct {
	users     map[string]GetUserByIDRow
	userRoles map[string]GetUserRoleRow
	admins    map[string]bool
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:     make(map[string]GetUserByIDRow),
		userRoles: make(map[string]GetUserRoleRow),
		admins:    make(map[string]bool),
	}
}

// Helper methods for setting up test data
func (m *MockRepository) AddUser(user GetUserByIDRow) {
	m.users[user.ID.String()] = user
}

func (m *MockRepository) AddUserRole(role GetUserRoleRow) {
	m.userRoles[role.UserID.String()] = role
}

func (m *MockRepository) SetAdmin(userID string, isAdmin bool) {
	m.admins[userID] = isAdmin
}

// Implementation of Repository interface
func (m *MockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (GetUserByIDRow, error) {
	user, exists := m.users[id.String()]
	if !exists {
		return GetUserByIDRow{}, ErrRecordNotFound
	}
	return user, nil
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (GetUserByEmailRow, error) {
	// Return mock data
	return GetUserByEmailRow{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "$2a$10$YZjEaHHtUBD/4RniGrx7ZO5TQShEBurJmc4Yz9Un.RFS4rP1W1hjm",
	}, nil
}

func (m *MockRepository) UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error) {
	now := time.Now()
	return User{
		ID:        arg.ID,
		Name:      arg.Name,
		Email:     arg.Email,
		CreatedAt: now,
		UpdatedAt: &now,
	}, nil
}

func (m *MockRepository) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) (uuid.UUID, error) {
	return arg.ID, nil
}

func (m *MockRepository) GetUserStats(ctx context.Context, id uuid.UUID) (GetUserStatsRow, error) {
	return GetUserStatsRow{
		ID:                  id,
		TotalIncomeRecords:  10,
		TotalExpenseRecords: 5,
		TotalBudgets:        2,
	}, nil
}

func (m *MockRepository) GetUserRole(ctx context.Context, id uuid.UUID) (GetUserRoleRow, error) {
	role, exists := m.userRoles[id.String()]
	if !exists {
		return GetUserRoleRow{}, ErrRecordNotFound
	}
	return role, nil
}

func (m *MockRepository) CheckUserIsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.admins[userID.String()], nil
}

func (m *MockRepository) UpdateUserRole(ctx context.Context, arg UpdateUserRoleParams) ([]byte, error) {
	return []byte(`{"user_id":"` + arg.UserID.String() + `","role_id":"` + arg.RoleID.String() + `"}`), nil
}

func (m *MockRepository) ListUsersByRole(ctx context.Context, name string) ([]ListUsersByRoleRow, error) {
	return []ListUsersByRoleRow{
		{
			UID:    uuid.New(),
			UName:  "Test User",
			UEmail: "test@example.com",
			Role:   name,
		},
	}, nil
}
