package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jorge-dev/centsible/internal/auth"
	"github.com/jorge-dev/centsible/internal/repository"
	"github.com/jorge-dev/centsible/server/handlers"
	customMiddleware "github.com/jorge-dev/centsible/server/middleware"
)

func (s *Server) RegisterRoutes(conn *pgx.Conn, jwtManager auth.JWTManager, env string) http.Handler {

	queries := repository.New(conn)
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(middleware.Recoverer)

	// Public routes
	r.Group(func(r chi.Router) {
		r.Get("/live", s.liveCheck)
		r.Get("/health", s.healthHandler)

		// Add seed routes (only in development)
		if env == "local" {
			seedHandler := handlers.NewSeedHandler(queries)
			r.Post("/api/seed", seedHandler.SeedData)
			r.Delete("/api/seed", seedHandler.DeleteSeedData)
		}

	})

	// Auth routes
	r.Group(func(r chi.Router) {
		authHandler := handlers.NewAuthHandler(queries, &jwtManager)
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Private routes
	r.Group(func(r chi.Router) {
		authMiddleware := customMiddleware.NewAuthMiddleware(&jwtManager)
		r.Use(authMiddleware.AuthRequired)

		// User routes
		userHandler := handlers.NewUserHandler(queries)
		r.Get("/user/profile", userHandler.GetProfile)
		r.Put("/user/profile", userHandler.UpdateProfile)
		r.Put("/user/password", userHandler.UpdatePassword)
		r.Get("/user/stats", userHandler.GetStats)

		// Income routes
		incomeHandler := handlers.NewIncomeHandler(queries)
		r.Post("/income", incomeHandler.CreateIncome)
		r.Get("/income", incomeHandler.GetIncomeList)
		r.Get("/income/{id}", incomeHandler.GetIncomeByID)
		r.Put("/income/{id}", incomeHandler.UpdateIncome)
		r.Delete("/income/{id}", incomeHandler.DeleteIncome)

		// Expense routes
		expenseHandler := handlers.NewExpenseHandler(queries)
		r.Post("/expenses", expenseHandler.CreateExpense)
		r.Get("/expenses", expenseHandler.ListExpenses)
		r.Get("/expenses/{id}", expenseHandler.GetExpenseByID)
		r.Put("/expenses/{id}", expenseHandler.UpdateExpense)
		r.Delete("/expenses/{id}", expenseHandler.DeleteExpense)
		r.Get("/expenses/category/{category}", expenseHandler.GetExpensesByCategory)
		r.Get("/expenses/range", expenseHandler.GetExpensesByDateRange)
		// Add missing routes
		r.Get("/expenses/monthly/total", expenseHandler.GetMonthlyExpenseTotal)
		r.Get("/expenses/category/totals", expenseHandler.GetExpenseTotalsByCategory)
		r.Get("/expenses/recent", expenseHandler.GetRecentExpenses)

		// Category routes
		categoryHandler := handlers.NewCategoryHandler(queries)
		r.Post("/categories", categoryHandler.CreateCategory)
		r.Get("/categories", categoryHandler.ListCategories)
		r.Get("/categories/{id}", categoryHandler.GetCategory)
		r.Put("/categories/{id}", categoryHandler.UpdateCategory)
		r.Delete("/categories/{id}", categoryHandler.DeleteCategory)
		r.Get("/categories/{id}/stats", categoryHandler.GetCategoryStats)
		r.Get("/categories/stats/most-used", categoryHandler.GetMostUsedCategories)
	})

	return r
}

func (s *Server) liveCheck(w http.ResponseWriter, r *http.Request) {

	liveCheckResponse := map[string]string{"status": "ok", "version": "0.0.1", "message": "Service is running"}

	if err := writeJSON(w, http.StatusOK, liveCheckResponse); err != nil {
		log.Printf("Error writing JSON: %v", err)
	}

}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	healthResponse := s.db.Health()
	if err := writeJSON(w, http.StatusOK, healthResponse); err != nil {
		log.Printf("Error writing JSON: %v", err)
	}

}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
