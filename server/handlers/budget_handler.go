package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/middleware"
)

type BudgetHandler struct {
	db *repository.Queries
}

type CreateBudgetRequest struct {
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	CategoryID uuid.UUID `json:"category_id"`
	Type       string    `json:"type"` // "recurring" or "one-time"
	StartDate  string    `json:"start_date"`
	EndDate    string    `json:"end_date"`
}

func NewBudgetHandler(db *repository.Queries) *BudgetHandler {
	return &BudgetHandler{db: db}
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		http.Error(w, "Invalid end date format", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

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
	bid, err := uuid.Parse(budgetID)
	if err != nil {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
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
	log.Println(alertThreshold)
	// Default to 80% if no threshold is provided
	if alertThreshold == "" {
		alertThreshold = "80"
	}

	// convert alert threshold to float
	threshold, err := strconv.ParseFloat(alertThreshold, 64)
	if err != nil {
		http.Error(w, "Invalid alert threshold. Must be a number", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
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
	uid, err := uuid.Parse(userID)
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
	bid, err := uuid.Parse(budgetID)
	if err != nil {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get current budget data
	currentBudget, err := h.db.GetBudgetByID(r.Context(), repository.GetBudgetByIDParams{
		ID:     bid,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Budget not found", http.StatusNotFound)
		return
	}

	// Use current values if request fields are empty/zero
	amount := currentBudget.Amount
	currency := currentBudget.Currency
	categoryID := currentBudget.CategoryID
	budgetType := currentBudget.Type
	startDate := currentBudget.StartDate
	endDate := currentBudget.EndDate

	if req.Amount != 0 {
		amount = req.Amount
	}
	if req.Currency != "" {
		currency = req.Currency
	}
	if req.CategoryID != uuid.Nil {
		categoryID = req.CategoryID
	}
	if req.Type != "" {
		budgetType = req.Type
	}
	if req.StartDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			http.Error(w, "Invalid start date format", http.StatusBadRequest)
			return
		}
		startDate = parsedDate
	}
	if req.EndDate != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		endDate = parsedDate
	}

	budget, err := h.db.UpdateBudget(r.Context(), repository.UpdateBudgetParams{
		ID:         bid,
		UserID:     uid,
		Amount:     amount,
		Currency:   currency,
		CategoryID: categoryID,
		Type:       budgetType,
		StartDate:  startDate,
		EndDate:    endDate,
	})
	if err != nil {
		http.Error(w, "Error updating budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	budgetID := chi.URLParam(r, "id")
	bid, err := uuid.Parse(budgetID)
	if err != nil {
		http.Error(w, "Invalid budget ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
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
	uid, err := uuid.Parse(userID)
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

func (h *BudgetHandler) GetBudgetsByCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "categoryId")
	cid, err := uuid.Parse(categoryID)
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
