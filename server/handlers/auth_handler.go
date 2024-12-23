package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/internal/validation"
)

type AuthHandler struct {
	db         repository.Repository
	jwtManager *auth.JWTManager
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthResponse struct {
	User      AuthUser       `json:"user"`
	TokenPair auth.TokenPair `json:"tokens"`
}

func NewAuthHandler(db repository.Repository, jm *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		db:         db,
		jwtManager: jm,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request with IsLogin = false for registration
	validator := &validation.AuthValidation{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		IsLogin:  false, // This is a registration request
	}
	if err := validator.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if email already exists
	emailExists, err := h.db.CheckEmailExists(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}
	if emailExists {
		http.Error(w, "An account with this email already exists. Please login or use a different email.", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Create user
	userID := uuid.New()
	user, err := h.db.CreateUser(r.Context(), repository.CreateUserParams{
		ID:           userID,
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Validate role ID before generating token
	if _, err := validation.ValidateRole(user.RoleID.String()); err != nil {
		http.Error(w, "Invalid role assigned", http.StatusInternalServerError)
		return
	}

	// Generate JWT pair
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID.String(), user.Email, user.RoleID.String())
	if err != nil {
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		TokenPair: *tokenPair,
		User: AuthUser{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request with IsLogin = true for login
	validator := &validation.AuthValidation{
		Email:    req.Email,
		Password: req.Password,
		IsLogin:  true, // This is a login request
	}
	if err := validator.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Validate password
	if !auth.ValidatePassword(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Validate role ID before generating token
	if _, err := validation.ValidateRole(user.RoleID.String()); err != nil {
		http.Error(w, "Invalid role assigned", http.StatusInternalServerError)
		return
	}

	// Generate JWT pair
	tokenPair, err := h.jwtManager.GenerateTokenPair(user.ID.String(), user.Email, user.RoleID.String())
	if err != nil {
		http.Error(w, "Error generating tokens", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{
		TokenPair: *tokenPair,
		User: AuthUser{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) Signout(w http.ResponseWriter, r *http.Request) {
	// Get token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "No token provided", http.StatusBadRequest)
		return
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	// Invalidate the token
	if err := h.jwtManager.InvalidateToken(token); err != nil {
		http.Error(w, "Error invalidating token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Successfully signed out",
	})
}
