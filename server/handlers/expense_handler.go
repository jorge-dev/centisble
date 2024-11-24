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

type ExpenseHandler struct {
	db repository.Repository
}

type ExpenseRequest struct {
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	CategoryID  uuid.UUID `json:"category_id"`
	Date        string    `json:"date"`
	Description string    `json:"description"`
}

func NewExpenseHandler(db repository.Repository) *ExpenseHandler {
	return &ExpenseHandler{db: db}
}

// CreateExpense handles POST /expenses
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req ExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	date, err := time.Parse(time.RFC3339, req.Date)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	expense, err := h.db.CreateExpense(r.Context(), repository.CreateExpenseParams{
		ID:          uuid.New(),
		UserID:      uid,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CategoryID:  req.CategoryID,
		Date:        date,
		Description: req.Description,
	})
	if err != nil {
		http.Error(w, "Error creating expense", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

// GetExpenseByID handles GET /expenses/{id}
func (h *ExpenseHandler) GetExpenseByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	expenseID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	expense, err := h.db.GetExpenseByID(r.Context(), repository.GetExpenseByIDParams{
		ID:     expenseID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

// ListExpenses handles GET /expenses
func (h *ExpenseHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	expenses, err := h.db.ListExpenses(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error fetching expenses", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

// GetExpensesByCategory handles GET /expenses/category/{category}
func (h *ExpenseHandler) GetExpensesByCategory(w http.ResponseWriter, r *http.Request) {
	categoryID := chi.URLParam(r, "category")
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

	expenses, err := h.db.GetExpensesByCategory(r.Context(), repository.GetExpensesByCategoryParams{
		UserID:     uid,
		CategoryID: cid,
	})
	if err != nil {
		http.Error(w, "Error fetching expenses", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

// GetExpensesByDateRange handles GET /expenses/range
func (h *ExpenseHandler) GetExpensesByDateRange(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	start, err := time.Parse(time.RFC3339, startDate)
	if err != nil {
		http.Error(w, "Invalid start date", http.StatusBadRequest)
		return
	}

	end, err := time.Parse(time.RFC3339, endDate)
	if err != nil {
		http.Error(w, "Invalid end date", http.StatusBadRequest)
		return
	}

	// Ensure start date is before end date
	if start.After(end) {
		http.Error(w, "Start date must be before end date", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	expenses, err := h.db.GetExpensesByDateRange(r.Context(), repository.GetExpensesByDateRangeParams{
		UserID:    uid,
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		http.Error(w, "Error fetching expenses", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

// UpdateExpense handles PUT /expenses/{id}
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	var req ExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id := chi.URLParam(r, "id")
	expenseID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Get current expense data
	currentExpense, err := h.db.GetExpenseByID(r.Context(), repository.GetExpenseByIDParams{
		ID:     expenseID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	// Use current values if request fields are empty/zero
	amount := currentExpense.Amount
	currency := currentExpense.Currency
	categoryID := currentExpense.CategoryID
	date := currentExpense.Date
	description := currentExpense.Description

	if req.Amount != 0 {
		amount = req.Amount
	}
	if req.Currency != "" {
		currency = req.Currency
	}
	if req.CategoryID != uuid.Nil {
		categoryID = req.CategoryID
	}
	if req.Date != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.Date)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		date = parsedDate
	}
	if req.Description != "" {
		description = req.Description
	}

	expense, err := h.db.UpdateExpense(r.Context(), repository.UpdateExpenseParams{
		ID:          expenseID,
		Amount:      amount,
		Currency:    currency,
		CategoryID:  categoryID,
		Date:        date,
		Description: description,
		UserID:      uid,
	})
	if err != nil {
		http.Error(w, "Error updating expense", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expense)
}

// DeleteExpense handles DELETE /expenses/{id}
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	expenseID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.DeleteExpense(r.Context(), repository.DeleteExpenseParams{
		ID:     expenseID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Error deleting expense", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// GetMonthlyExpenseTotal handles GET /expenses/monthly/total
func (h *ExpenseHandler) GetMonthlyExpenseTotal(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date-time")
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	totals, err := h.db.GetMonthlyExpenseTotal(r.Context(), repository.GetMonthlyExpenseTotalParams{
		UserID: uid,
		Date:   date,
	})
	if err != nil {
		http.Error(w, "Error fetching monthly totals", http.StatusInternalServerError)
		return
	}

	if totals == nil {
		totals = []repository.GetMonthlyExpenseTotalRow{}

	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totals)
}

// GetExpenseTotalsByCategory handles GET /expenses/category/totals
func (h *ExpenseHandler) GetExpenseTotalsByCategory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	totals, err := h.db.GetExpenseTotalsByCategory(r.Context(), uid)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error fetching category totals", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totals)
}

// GetRecentExpenses handles GET /expenses/recent
func (h *ExpenseHandler) GetRecentExpenses(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.ParseInt(limitStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	expenses, err := h.db.GetRecentExpenses(r.Context(), repository.GetRecentExpensesParams{
		UserID: uid,
		Limit:  int32(limit),
	})
	if err != nil {
		http.Error(w, "Error fetching recent expenses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}
