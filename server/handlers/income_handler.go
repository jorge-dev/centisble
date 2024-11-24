package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/middleware"
)

type IncomeHandler struct {
	db repository.Repository
}

func NewIncomeHandler(db repository.Repository) *IncomeHandler {
	return &IncomeHandler{db: db}
}

type CreateIncomeRequest struct {
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Source      string    `json:"source"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
}

func (h *IncomeHandler) CreateIncome(w http.ResponseWriter, r *http.Request) {
	var req CreateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err.Error())
		errMsg := "Invalid request body: "
		switch {
		case err.Error() == "EOF":
			errMsg += "request body is empty"
		case err.Error() == "unexpected end of JSON input":
			errMsg += "incomplete JSON data"
		default:
			errMsg += "malformed JSON - " + err.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Amount <= 0 {
		http.Error(w, "Invalid request body: amount must be greater than 0", http.StatusBadRequest)
		return
	}
	if req.Currency == "" {
		http.Error(w, "Invalid request body: currency is required", http.StatusBadRequest)
		return
	}
	if req.Source == "" {
		http.Error(w, "Invalid request body: source is required", http.StatusBadRequest)
		return
	}
	if req.Date.IsZero() {
		http.Error(w, "Invalid request body: date is required", http.StatusBadRequest)
		return
	}

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

	income, err := h.db.CreateIncome(r.Context(), repository.CreateIncomeParams{
		ID:          uuid.New(),
		UserID:      uid,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Source:      req.Source,
		Date:        req.Date,
		Description: req.Description,
	})
	if err != nil {
		http.Error(w, "Error creating income record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(income)
}

func (h *IncomeHandler) GetIncomeList(w http.ResponseWriter, r *http.Request) {
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

	incomes, err := h.db.ListIncome(r.Context(), uid)
	if err != nil {
		http.Error(w, "Error fetching income records", http.StatusInternalServerError)
		return
	}

	//TODO: find a better way to handle empty slices
	if len(incomes) == 0 {
		incomes = []repository.Income{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incomes)
}

func (h *IncomeHandler) GetIncomeByID(w http.ResponseWriter, r *http.Request) {
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

	incomeID := chi.URLParam(r, "id")
	incomeUUID, err := uuid.Parse(incomeID)
	if err != nil {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	income, err := h.db.GetIncomeByID(r.Context(), repository.GetIncomeByIDParams{
		ID:     incomeUUID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Income record not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(income)
}

func (h *IncomeHandler) UpdateIncome(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg := "Invalid request body: "
		switch {
		case err.Error() == "EOF":
			errMsg += "request body is empty"
		case err.Error() == "unexpected end of JSON input":
			errMsg += "incomplete JSON data"
		default:
			errMsg += "malformed JSON - " + err.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if len(req) == 0 {
		http.Error(w, "Invalid request body: at least one field must be provided for update", http.StatusBadRequest)
		return
	}

	// Get existing income record first
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, _ := uuid.Parse(userID)
	incomeID := chi.URLParam(r, "id")
	incomeUUID, _ := uuid.Parse(incomeID)

	currentIncome, err := h.db.GetIncomeByID(r.Context(), repository.GetIncomeByIDParams{
		ID:     incomeUUID,
		UserID: uid,
	})
	if err != nil {
		http.Error(w, "Income record not found", http.StatusNotFound)
		return
	}

	updateParams := repository.UpdateIncomeParams{
		ID:          incomeUUID,
		UserID:      uid,
		Amount:      currentIncome.Amount,
		Currency:    currentIncome.Currency,
		Source:      currentIncome.Source,
		Date:        currentIncome.Date,
		Description: currentIncome.Description,
	}

	// Only validate and update fields that are present in the request
	if amount, ok := req["amount"].(float64); ok {
		if amount <= 0 {
			http.Error(w, "Invalid request body: amount must be greater than 0", http.StatusBadRequest)
			return
		}
		updateParams.Amount = amount
	}

	if currency, ok := req["currency"].(string); ok {
		if currency == "" {
			http.Error(w, "Invalid request body: currency cannot be empty", http.StatusBadRequest)
			return
		}
		updateParams.Currency = currency
	}

	if source, ok := req["source"].(string); ok {
		if source == "" {
			http.Error(w, "Invalid request body: source cannot be empty", http.StatusBadRequest)
			return
		}
		updateParams.Source = source
	}

	if dateStr, ok := req["date"].(string); ok {
		date, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			http.Error(w, "Invalid request body: invalid date format", http.StatusBadRequest)
			return
		}
		updateParams.Date = date
	}

	if description, ok := req["description"].(string); ok {
		updateParams.Description = description
	}

	income, err := h.db.UpdateIncome(r.Context(), updateParams)
	if err != nil {
		http.Error(w, "Error updating income record", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(income)
}

func (h *IncomeHandler) DeleteIncome(w http.ResponseWriter, r *http.Request) {
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

	incomeID := chi.URLParam(r, "id")
	incomeUUID, err := uuid.Parse(incomeID)
	if err != nil {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	rows, err := h.db.DeleteIncome(r.Context(), repository.DeleteIncomeParams{
		ID:     incomeUUID,
		UserID: uid,
	})

	if err != nil {
		http.Error(w, "Error deleting income record", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		http.Error(w, "Income record not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
