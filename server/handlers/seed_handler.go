package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jorge-dev/centsible/internal/repository"
)

type SeedHandler struct {
	queries *repository.Queries
}

func NewSeedHandler(queries *repository.Queries) *SeedHandler {
	return &SeedHandler{
		queries: queries,
	}
}

func (h *SeedHandler) SeedData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Execute seed operations in sequence
	if err := h.queries.SeedUsers(ctx); err != nil {
		log.Printf("Error seeding users: %v", err)
		http.Error(w, "Error seeding database", http.StatusInternalServerError)
		return
	}

	if err := h.queries.SeedCategories(ctx); err != nil {
		log.Printf("Error seeding categories: %v", err)
		http.Error(w, "Error seeding database", http.StatusInternalServerError)
		return
	}

	if err := h.queries.SeedIncome(ctx); err != nil {
		log.Printf("Error seeding income: %v", err)
		http.Error(w, "Error seeding database", http.StatusInternalServerError)
		return
	}

	if err := h.queries.SeedExpenses(ctx); err != nil {
		log.Printf("Error seeding expenses: %v", err)
		http.Error(w, "Error seeding database", http.StatusInternalServerError)
		return
	}

	if err := h.queries.SeedBudgets(ctx); err != nil {
		log.Printf("Error seeding budgets: %v", err)
		http.Error(w, "Error seeding database", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Database seeded successfully",
	})
}

func (h *SeedHandler) DeleteSeedData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if err := h.queries.DeleteSeedData(ctx); err != nil {
		log.Printf("Error deleting seed data: %v", err)
		http.Error(w, "Error deleting seed data", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Seed data deleted successfully",
	})
}
