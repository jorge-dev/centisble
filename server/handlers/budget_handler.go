package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/internal/validation"
	"github.com/jorge-dev/centsible/server/middleware"
)

type BudgetHandler struct {
	db repository.Repository
}

type CreateBudgetRequest struct {
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	CategoryID uuid.UUID `json:"category_id"`
	Type       string    `json:"type"` // "recurring" or "one-time"
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
}

func NewBudgetHandler(db repository.Repository) *BudgetHandler {
	return &BudgetHandler{db: db}
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	validator := &validation.BudgetValidation{
		Amount:     req.Amount,
		Currency:   req.Currency,
		CategoryID: req.CategoryID,
		Type:       req.Type,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
	}

	if err := validator.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	startDate, _ := validation.ValidateDate(req.StartDate)
	endDate, _ := validation.ValidateDate(req.EndDate)

	budget, err := h.db.CreateBudget(r.Context(), repository.CreateBudgetParams{
		ID:         uuid.New(),
		UserID:     uid,
		Amount:     req.Amount,
		Currency:   req.Currency,
		CategoryID: req.CategoryID,
		Type:       req.Type,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		http.Error(w, "Error creating budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) GetBudgetUsage(w http.ResponseWriter, r *http.Request) {
	budgetID := chi.URLParam(r, "id")
	bid, err := validation.ValidateUUID(budgetID)
	if err != nil {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	usage, err := h.db.GetBudgetUsage(r.Context(), repository.GetBudgetUsageParams{
		BudgetID: bid,
		UserID:   uid,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting budget usage", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usage)
}

func (h *BudgetHandler) GetBudgetsNearLimit(w http.ResponseWriter, r *http.Request) {
	alertThreshold := r.URL.Query().Get("alert_threshold")
	threshold := 80.0 // Default value

	if alertThreshold != "" {
		var err error
		threshold, err = strconv.ParseFloat(alertThreshold, 64)
		if err != nil {
			http.Error(w, "Invalid alert threshold. Must be a number", http.StatusBadRequest)
			return
		}

		validator := &validation.BudgetValidation{
			AlertThreshold: &threshold,
		}
		if err := validator.ValidateAlertThreshold(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get budgets that are at 80% or more of their limit
	alerts, err := h.db.GetBudgetsNearLimit(r.Context(), repository.GetBudgetsNearLimitParams{
		UserID:    uid,
		Threshold: threshold, // Alert threshold percentage
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Error getting budget alerts", http.StatusInternalServerError)
		return
	}

	if len(alerts) == 0 {
		alerts = []repository.GetBudgetsNearLimitRow{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func (h *BudgetHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	budgets, err := h.db.ListBudgets(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error listing budgets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	budgetID := chi.URLParam(r, "id")
	bid, err := validation.ValidateUUID(budgetID)
	if err != nil {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	currentBudget, err := h.db.GetBudgetByID(r.Context(), repository.GetBudgetByIDParams{
		ID:     bid,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Budget not found", http.StatusNotFound)
		return
	}

	validator := &validation.BudgetValidation{
		Amount:          req.Amount,
		Currency:        req.Currency,
		CategoryID:      req.CategoryID,
		Type:            req.Type,
		StartDate:       req.StartDate,
		EndDate:         req.EndDate,
		IsPartialUpdate: true, // Set this flag for updates
	}

	if err := validator.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	current := validation.CurrentBudget{
		Amount:     currentBudget.Amount,
		Currency:   currentBudget.Currency,
		CategoryID: currentBudget.CategoryID,
		Type:       currentBudget.Type,
		StartDate:  currentBudget.StartDate,
		EndDate:    currentBudget.EndDate,
	}

	validated, err := validator.ValidatePartialUpdate(current)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	budget, err := h.db.UpdateBudget(r.Context(), repository.UpdateBudgetParams{
		ID:         bid,
		UserID:     uid,
		Amount:     validated.Amount,
		Currency:   validated.Currency,
		CategoryID: validated.CategoryID,
		Type:       validated.Type,
		StartDate:  validated.StartDate,
		EndDate:    validated.EndDate,
	})
	if err != nil {
		log.Println(err)
		http.Error(w, "Error updating budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	budgetID := chi.URLParam(r, "id")
	bid, err := validation.ValidateUUID(budgetID)
	if err != nil {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rows, err := h.db.DeleteBudget(r.Context(), repository.DeleteBudgetParams{
		ID:     bid,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Error deleting budget", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "Budget not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BudgetHandler) GetRecurringBudgets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	budgets, err := h.db.GetRecurringBudgets(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error getting recurring budgets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

func (h *BudgetHandler) GetOneTimeBudgets(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	budgets, err := h.db.GetOneTimeBudgets(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error getting one-time budgets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}

func (h *BudgetHandler) GetBudgetsByCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")
	cid, err := validation.ValidateUUID(categoryID)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := validation.ValidateUUID(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	budgets, err := h.db.GetBudgetsByCategory(r.Context(), repository.GetBudgetsByCategoryParams{
		UserID:     uid,
		CategoryID: cid,
	})
	if err != nil {
		http.Error(w, "Error getting budgets by category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budgets)
}
