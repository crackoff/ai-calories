package main

import (
	"ai-calories/ai"
	"ai-calories/database"
	"ai-calories/internal/config"
	"ai-calories/internal/handler"
	"ai-calories/internal/model"
	"ai-calories/internal/repository"
	"ai-calories/internal/service"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	cfg := config.Load()

	// Database
	db := database.NewDatabase(cfg.DatabaseURL)
	gormDB := db.DB
	gormDB.AutoMigrate(&model.RefreshToken{}, &model.FoodCache{}, &model.PaymentHistory{})

	// AI Classifier
	classifier := ai.NewClassifier("openai", "food")

	// Repositories
	userRepo := repository.NewUserRepository(gormDB)
	refreshTokenRepo := repository.NewRefreshTokenRepository(gormDB)
	foodRepo := repository.NewFoodRepository(gormDB)
	foodCacheRepo := repository.NewFoodCacheRepository(gormDB)
	paymentRepo := repository.NewPaymentRepository(gormDB)

	// Services
	authService := service.NewAuthService(userRepo, refreshTokenRepo, cfg)
	foodService := service.NewFoodService(foodRepo, foodCacheRepo, userRepo, classifier)
	foodCacheService := service.NewFoodCacheService(foodCacheRepo)
	userService := service.NewUserService(userRepo)
	paymentService := service.NewPaymentService(paymentRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	foodHandler := handler.NewFoodHandler(foodService)
	foodCacheHandler := handler.NewFoodCacheHandler(foodCacheService)
	userHandler := handler.NewUserHandler(userService)
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Router
	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Route("/api/v1", func(r chi.Router) {
		// Public auth routes
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/google", authHandler.GoogleLogin)
		r.Post("/auth/apple", authHandler.AppleLogin)
		r.Post("/auth/refresh", authHandler.Refresh)

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(handler.AuthMiddleware(cfg.JWTSecret))

			// Food
			r.Post("/food", foodHandler.LogFood)
			r.Get("/food/today", foodHandler.GetTodayFoods)
			r.Get("/food/date/{date}", foodHandler.GetFoodsByDate)
			r.Get("/food/summary/today", foodHandler.GetTodaySummary)
			r.Get("/food/summary/{date}", foodHandler.GetDateSummary)
			r.Get("/food/history", foodHandler.GetFoodHistory)
			r.Delete("/food/last", foodHandler.DeleteLast)
			r.Delete("/food/{id}", foodHandler.DeleteByID)

			// Food cache
			r.Get("/food-cache/search", foodCacheHandler.Search)
			r.Get("/food-cache/{id}", foodCacheHandler.GetByID)

			// User
			r.Get("/user/profile", userHandler.GetProfile)
			r.Put("/user/timezone", userHandler.UpdateTimezone)
			r.Put("/user/language", userHandler.UpdateLanguage)

			// Payments
			r.Get("/payments/current", paymentHandler.GetCurrent)
			r.Post("/payments", paymentHandler.Record)
			r.Get("/payments/history", paymentHandler.GetHistory)
		})
	})

	log.Printf("API server starting on :%s", cfg.APIPort)
	log.Fatal(http.ListenAndServe(":"+cfg.APIPort, r))
}
