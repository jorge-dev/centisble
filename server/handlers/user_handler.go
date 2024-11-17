package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/middleware"
)

type UserHandler struct {
	db *repository.Queries
}

type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func NewUserHandler(db *repository.Queries) *UserHandler {
	return &UserHandler{db: db}
}

// GetProfile handles GET /api/user/profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.db.GetUserByID(r.Context(), uid)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateProfile handles PUT /api/user/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get current user data
	currentUser, err := h.db.GetUserByID(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error fetching user data", http.StatusInternalServerError)
		return
	}

	// Use current values if request fields are empty
	name := currentUser.Name
	email := currentUser.Email

	if req.Name != "" {
		name = req.Name
	}
	if req.Email != "" {
		email = req.Email
	}

	user, err := h.db.UpdateUser(r.Context(), repository.UpdateUserParams{
		ID:    uid,
		Name:  name,
		Email: email,
	})
	if err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdatePassword handles PUT /api/user/password
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get current user to verify password
	user, err := h.db.GetUserByEmail(r.Context(), r.Context().Value(middleware.EmailKey).(string))
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Validate current password
	if !auth.ValidatePassword(req.CurrentPassword, user.PasswordHash) {
		http.Error(w, "Invalid current password", http.StatusUnauthorized)
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Error processing password", http.StatusInternalServerError)
		return
	}

	// Update password
	_, err = h.db.UpdateUserPassword(r.Context(), repository.UpdateUserPasswordParams{
		ID:           uid,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

}

// GetStats handles GET /api/user/stats
func (h *UserHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	stats, err := h.db.GetUserStats(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error fetching user stats", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetUserRole handles GET /api/user/role
func (h *UserHandler) GetUserRole(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	role, err := h.db.GetUserRole(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
}

// UpdateUserRole handles PUT /api/user/role
func (h *UserHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// check if the current user is an admin
	isAdmin, err := h.db.CheckUserIsAdmin(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error checking user role", http.StatusInternalServerError)
		return
	}

	if !isAdmin {
		http.Error(w, "You do not have permission to update user roles", http.StatusForbidden)
		return
	}

	// get the roleID from the request body
	var roleID struct {
		RoleID string `json:"role_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&roleID); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rId, err := uuid.Parse(roleID.RoleID)

	if err != nil {
		http.Error(w, "Invalid role ID", http.StatusBadRequest)
		return
	}

	roleInfo, err := h.db.UpdateUserRole(r.Context(), repository.UpdateUserRoleParams{
		RoleID: rId,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Error updating user role", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roleInfo)

}

// ListUsersByRole handles GET /api/users/role/list
func (h *UserHandler) ListUsersByRole(w http.ResponseWriter, r *http.Request) {
	// get role name from params
	roleName := r.URL.Query().Get("role_name")
	if roleName == "" {
		http.Error(w, "Role name is required", http.StatusBadRequest)
		return
	}

	users, err := h.db.ListUsersByRole(r.Context(), roleName)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		http.Error(w, "No users found for this role. Please make sure role is valid", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)

}
