package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fernando/financial-management-tracking/backend/internal/account"
	"github.com/fernando/financial-management-tracking/backend/internal/allocation"
	"github.com/fernando/financial-management-tracking/backend/internal/auth"
	"github.com/fernando/financial-management-tracking/backend/internal/budget"
	"github.com/fernando/financial-management-tracking/backend/internal/category"
	"github.com/fernando/financial-management-tracking/backend/internal/config"
	"github.com/fernando/financial-management-tracking/backend/internal/cycle"
	"github.com/fernando/financial-management-tracking/backend/internal/dashboard"
	"github.com/fernando/financial-management-tracking/backend/internal/database"
	"github.com/fernando/financial-management-tracking/backend/internal/health"
	"github.com/fernando/financial-management-tracking/backend/internal/investment"
	"github.com/fernando/financial-management-tracking/backend/internal/middleware"
	"github.com/fernando/financial-management-tracking/backend/internal/receivable"
	"github.com/fernando/financial-management-tracking/backend/internal/report"
	"github.com/fernando/financial-management-tracking/backend/internal/transaction"
	"github.com/go-chi/chi/v5"
)

func main() {
	// ── Logger ────────────────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── Config ────────────────────────────────────────────────────────
	cfg := config.Load()
	slog.Info("configuration loaded", "port", cfg.Port, "env", cfg.Env)

	// ── Database ──────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connected")

	// Run migrations
	if err := database.Migrate(ctx, pool, "migrations"); err != nil {
		slog.Error("failed to run database migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migrations verified")

	// ── Services & Handlers ──────────────────────────────────────────
	tokenService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTExpiry)
	authRepo := auth.NewRepository(pool)
	authService := auth.NewService(authRepo, tokenService)
	authHandler := auth.NewHandler(authService, tokenService)

	accountRepo := account.NewRepository(pool)
	accountService := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountService)

	categoryRepo := category.NewRepository(pool)
	categoryService := category.NewService(categoryRepo)
	categoryHandler := category.NewHandler(categoryService)

	transactionRepo := transaction.NewRepository(pool)
	transactionService := transaction.NewService(transactionRepo)
	transactionHandler := transaction.NewHandler(transactionService)

	cycleRepo := cycle.NewRepository(pool)
	cycleService := cycle.NewService(cycleRepo)
	cycleHandler := cycle.NewHandler(cycleService)

	budgetRepo := budget.NewRepository(pool)
	budgetService := budget.NewService(budgetRepo)
	budgetHandler := budget.NewHandler(budgetService)

	allocationRepo := allocation.NewRepository(pool)
	allocationService := allocation.NewService(allocationRepo)
	allocationHandler := allocation.NewHandler(allocationService)

	dashboardRepo := dashboard.NewRepository(pool)
	dashboardService := dashboard.NewService(dashboardRepo)
	dashboardHandler := dashboard.NewHandler(dashboardService)

	receivableRepo := receivable.NewRepository(pool)
	receivableService := receivable.NewService(receivableRepo)
	receivableHandler := receivable.NewHandler(receivableService)

	investmentRepo := investment.NewRepository(pool)
	investmentService := investment.NewService(investmentRepo)
	investmentHandler := investment.NewHandler(investmentService)

	reportRepo := report.NewRepository(pool)
	reportService := report.NewService(reportRepo)
	reportHandler := report.NewHandler(reportService)

	// ── Router ────────────────────────────────────────────────────────
	r := chi.NewRouter()

	// Rate limiter (300 req/min with burst 50 per client IP)
	rateLimiter := middleware.NewRateLimiter(300, 50)

	// Middleware stack
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.MaxBodySize(10 << 20)) // 10MB request size limit
	r.Use(middleware.CORS)
	r.Use(rateLimiter.Middleware())
	r.Use(middleware.Logger)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", health.Handler(pool))
		r.Mount("/auth", authHandler.Routes())

		// Authenticated domain routes
		r.Group(func(pr chi.Router) {
			pr.Use(auth.RequireAuth(tokenService))
			pr.Mount("/accounts", accountHandler.Routes())
			pr.Mount("/categories", categoryHandler.Routes())
			pr.Mount("/transactions", transactionHandler.Routes())
			pr.Mount("/cycle", cycleHandler.Routes())
			pr.Mount("/budgets", budgetHandler.Routes())
			pr.Mount("/allocations", allocationHandler.Routes())
			pr.Mount("/dashboard", dashboardHandler.Routes())
			pr.Mount("/receivables", receivableHandler.Routes())
			pr.Mount("/investments", investmentHandler.Routes())
			pr.Mount("/reports", reportHandler.Routes())
		})
	})

	// ── Server ────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("server shutting down")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}
