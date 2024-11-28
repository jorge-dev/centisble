package handlers

import (
	"encoding/json"
	"fmt"
	"log"
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

type TopCategory struct {
	CategoryID   string  `json:"category_id"`
	CategoryName string  `json:"category_name"`
	UsageCount   int     `json:"usage_count"`
	TotalSpent   float64 `json:"total_spent"`
}

type MonthlySummaryResponse struct {
	Currency      string        `json:"currency"`
	TotalIncome   float64       `json:"total_income"`
	TotalExpenses float64       `json:"total_expenses"`
	TotalSavings  float64       `json:"total_savings"`
	TopCategories []TopCategory `json:"top_categories"`
}

type MonthlyTrend struct {
	Month        time.Time `json:"month"`
	CategoryName string    `json:"category_name"`
	Amount       float64   `json:"amount"`
}

type YearlySummaryResponse struct {
	Currency      string         `json:"currency"`
	TotalIncome   float64        `json:"total_income"`
	TotalExpenses float64        `json:"total_expenses"`
	TotalSavings  float64        `json:"total_savings"`
	TopCategories []TopCategory  `json:"top_categories"`
	MonthlyTrend  []MonthlyTrend `json:"monthly_trend"`
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
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}
	log.Printf("date: %v", date)
	log.Printf("userId: %v", uid)

	summary, err := h.db.GetMonthlySummary(r.Context(), repository.GetMonthlySummaryParams{
		UserID: uid,
		Date:   date,
	})
	if err != nil {
		http.Error(w, "Error fetching monthly summary", http.StatusInternalServerError)
		return
	}
	if summary == nil {
		http.Error(w, fmt.Sprintf("No summary found for month %s", date.Format("January 2006")), http.StatusNotFound)
		return
	}

	var responses []MonthlySummaryResponse
	for _, s := range summary {
		var topCategories []TopCategory
		if err := json.Unmarshal(s.TopCategories, &topCategories); err != nil {
			http.Error(w, "Error processing summary data", http.StatusInternalServerError)
			return
		}

		responses = append(responses, MonthlySummaryResponse{
			Currency:      s.Currency,
			TotalIncome:   s.TotalIncome,
			TotalExpenses: s.TotalExpenses,
			TotalSavings:  s.TotalSavings,
			TopCategories: topCategories,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
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
		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
	}

	log.Printf("date: %s", date)
	log.Printf("userId: %s", uid)
	summary, err := h.db.GetYearlySummary(r.Context(), repository.GetYearlySummaryParams{
		UserID: uid,
		Date:   date,
	})
	if err != nil {
		http.Error(w, "Error fetching yearly summary", http.StatusInternalServerError)
		return
	}
	if summary == nil {
		http.Error(w, fmt.Sprintf("No summary found for year %s", date.Format("2006")), http.StatusNotFound)
		return
	}

	var responses []YearlySummaryResponse
	for _, s := range summary {
		var topCategories []TopCategory
		var monthlyTrend []MonthlyTrend

		if err := json.Unmarshal(s.TopCategories, &topCategories); err != nil {
			http.Error(w, "Error processing summary data", http.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(s.MonthlyTrend, &monthlyTrend); err != nil {

			http.Error(w, "Error processing trend data", http.StatusInternalServerError)
			return
		}

		responses = append(responses, YearlySummaryResponse{
			Currency:      s.Currency,
			TotalIncome:   s.TotalIncome,
			TotalExpenses: s.TotalExpenses,
			TotalSavings:  s.TotalSavings,
			TopCategories: topCategories,
			MonthlyTrend:  monthlyTrend,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}
