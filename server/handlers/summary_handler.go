package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/middleware"
)

type SummaryHandler struct {
	db repository.Repository
}

func NewSummaryHandler(db repository.Repository) *SummaryHandler {
	return &SummaryHandler{db: db}
}

// GetMonthlySummary handles GET /api/summary/monthly
func (h *SummaryHandler) GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse date from query params, default to current month if not provided
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	summary, err := h.db.GetMonthlySummary(r.Context(), repository.GetMonthlySummaryParams{
		UserID: uid,
		Date:   date,
	})
	if err != nil {
		http.Error(w, "Error fetching monthly summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

// GetYearlySummary handles GET /api/summary/yearly
func (h *SummaryHandler) GetYearlySummary(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	uid, err := uuid.Parse(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Parse date from query params, default to current year if not provided
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	summary, err := h.db.GetYearlySummary(r.Context(), repository.GetYearlySummaryParams{
		UserID: uid,
		Date:   date,
	})
	if err != nil {
		http.Error(w, "Error fetching yearly summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}
