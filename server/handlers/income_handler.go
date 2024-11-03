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
	db *repository.Queries
}

func NewIncomeHandler(db *repository.Queries) *IncomeHandler {
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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
	log.Printf("Incomes: %v", incomes)
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
	var req CreateIncomeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

	incomeID := chi.URLParam(r, "id")
	incomeUUID, err := uuid.Parse(incomeID)
	if err != nil {
		http.Error(w, "Invalid income ID", http.StatusBadRequest)
		return
	}

	income, err := h.db.UpdateIncome(r.Context(), repository.UpdateIncomeParams{
		ID:          incomeUUID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Source:      req.Source,
		Date:        req.Date,
		Description: req.Description,
		UserID:      uid,
	})
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
