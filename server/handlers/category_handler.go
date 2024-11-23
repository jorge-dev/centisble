package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/middleware"
)

type CategoryHandler struct {
	queries repository.Repository
}

func NewCategoryHandler(db repository.Repository) *CategoryHandler {
	return &CategoryHandler{queries: db}
}

type createCategoryRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req createCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// check if category is empty
	if req.Name == "" {
		http.Error(w, "Category name has to have a value", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	exists, err := h.queries.CheckCategoryExists(r.Context(), repository.CheckCategoryExistsParams{
		UserID: uid,
		Name:   req.Name,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Category already exists", http.StatusConflict)
		return
	}

	category, err := h.queries.CreateCategory(r.Context(), repository.CreateCategoryParams{
		ID:     uuid.New(),
		UserID: uid,
		Name:   req.Name,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	category, err := h.queries.GetCategoryByID(r.Context(), repository.GetCategoryByIDParams{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	categories, err := h.queries.ListCategories(r.Context(), uid)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, categories)
}

type updateCategoryRequest struct {
	Name string `json:"name"`
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req updateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	category, err := h.queries.UpdateCategory(r.Context(), repository.UpdateCategoryParams{
		ID:     id,
		Name:   req.Name,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	rows, err := h.queries.DeleteCategory(r.Context(), repository.DeleteCategoryParams{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CategoryHandler) GetCategoryStats(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}
	stats, err := h.queries.GetCategoryUsage(r.Context(), repository.GetCategoryUsageParams{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		log.Printf("Error getting category stats: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func (h *CategoryHandler) GetMostUsedCategories(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get limit from query params, default to 10
	limit := int32(10)
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit = int32(parsedLimit)
		}
	}

	stats, err := h.queries.GetMostUsedCategories(r.Context(), repository.GetMostUsedCategoriesParams{
		ID:    uid,
		Limit: limit,
	})
	if err != nil {
		log.Printf("Error getting most used categories: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
