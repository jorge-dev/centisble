package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
)

type UserMock struct {
	users     map[string]repository.GetUserByIDRow
	userRoles map[string]repository.GetUserRoleRow
	admins    map[string]bool
}

func NewUserMock() *UserMock {
	return &UserMock{
		users:     make(map[string]repository.GetUserByIDRow),
		userRoles: make(map[string]repository.GetUserRoleRow),
		admins:    make(map[string]bool),
	}
}

// Helper methods for setting up test data
func (m *UserMock) AddUser(user repository.GetUserByIDRow) {
	m.users[user.ID.String()] = user
}

func (m *UserMock) AddUserRole(role repository.GetUserRoleRow) {
	m.userRoles[role.UserID.String()] = role
}

func (m *UserMock) SetAdmin(userID string, isAdmin bool) {
	m.admins[userID] = isAdmin
}

// Implementation of Repository interface
func (m *UserMock) GetUserByID(ctx context.Context, id uuid.UUID) (repository.GetUserByIDRow, error) {
	user, exists := m.users[id.String()]
	if !exists {
		return repository.GetUserByIDRow{}, ErrRecordNotFound
	}
	return user, nil
}

func (m *UserMock) GetUserByEmail(ctx context.Context, email string) (repository.GetUserByEmailRow, error) {
	// Return mock data
	return repository.GetUserByEmailRow{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: "$2a$10$YZjEaHHtUBD/4RniGrx7ZO5TQShEBurJmc4Yz9Un.RFS4rP1W1hjm",
	}, nil
}

func (m *UserMock) UpdateUser(ctx context.Context, arg repository.UpdateUserParams) (repository.User, error) {
	now := time.Now()
	return repository.User{
		ID:        arg.ID,
		Name:      arg.Name,
		Email:     arg.Email,
		CreatedAt: now,
		UpdatedAt: &now,
	}, nil
}

func (m *UserMock) UpdateUserPassword(ctx context.Context, arg repository.UpdateUserPasswordParams) (uuid.UUID, error) {
	return arg.ID, nil
}

func (m *UserMock) GetUserStats(ctx context.Context, id uuid.UUID) (repository.GetUserStatsRow, error) {
	return repository.GetUserStatsRow{
		ID:                  id,
		TotalIncomeRecords:  10,
		TotalExpenseRecords: 5,
		TotalBudgets:        2,
	}, nil
}

func (m *UserMock) GetUserRole(ctx context.Context, id uuid.UUID) (repository.GetUserRoleRow, error) {
	role, exists := m.userRoles[id.String()]
	if !exists {
		return repository.GetUserRoleRow{}, ErrRecordNotFound
	}
	return role, nil
}

func (m *UserMock) CheckUserIsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.admins[userID.String()], nil
}

func (m *UserMock) UpdateUserRole(ctx context.Context, arg repository.UpdateUserRoleParams) ([]byte, error) {
	return []byte(`{"user_id":"` + arg.UserID.String() + `","role_id":"` + arg.RoleID.String() + `"}`), nil
}

func (m *UserMock) ListUsersByRole(ctx context.Context, name string) ([]repository.ListUsersByRoleRow, error) {
	return []repository.ListUsersByRoleRow{
		{
			UID:    uuid.New(),
			UName:  "Test User",
			UEmail: "test@example.com",
			Role:   name,
		},
	}, nil
}
