package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/handlers"
	appmiddleware "ticket-booking/backend/internal/middleware"
)

func main() {
	cfg := config.Load()
	log.Printf("Starting Ticket Booking Service in %s mode...", cfg.Environment)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to database
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("⚠️  Ensure PostgreSQL is running at %s", cfg.DatabaseURL)
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.Close()

	// Initialize SQLC queries and HTTP handlers
	queries := generated.New(database.Pool)
	authHandler := handlers.NewAuthHandler(queries, cfg)

	r := chi.NewRouter()

	// Global Middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health Check & Base Routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"status":    "ok",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	r.Get("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		handlers.RespondJSON(w, http.StatusOK, map[string]interface{}{
			"message": "Ticket Booking API v1",
			"status":  "running",
		})
	})

	// API v1 Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth Routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			// Protected /me endpoint
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.Authenticate(cfg.JWTSecret))
				r.Get("/me", authHandler.Me)
			})
		})

		// Sample Protected RBAC Route Groups
		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(cfg.JWTSecret))

			// Admin-only subroutes
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.AdminOnly)
				r.Get("/admin/ping", func(w http.ResponseWriter, r *http.Request) {
					handlers.RespondSuccess(w, http.StatusOK, map[string]string{"role": "ADMIN"}, "Admin access verified")
				})
			})

			// Organiser-only subroutes
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.OrganiserOnly)
				r.Get("/organiser/ping", func(w http.ResponseWriter, r *http.Request) {
					handlers.RespondSuccess(w, http.StatusOK, map[string]string{"role": "ORGANISER"}, "Organiser access verified")
				})
			})

			// Customer-only subroutes
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.CustomerOnly)
				r.Get("/customer/ping", func(w http.ResponseWriter, r *http.Request) {
					handlers.RespondSuccess(w, http.StatusOK, map[string]string{"role": "CUSTOMER"}, "Customer access verified")
				})
			})
		})
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("🚀 HTTP Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
