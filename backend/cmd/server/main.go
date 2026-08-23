package main

import (
	"context"
	"errors"
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
	venueHandler := handlers.NewVenueHandler(queries, database.Pool)
	eventHandler := handlers.NewEventHandler(queries, database.Pool)
	reservationHandler := handlers.NewReservationHandler(queries, database.Pool, cfg)
	bookingHandler := handlers.NewBookingHandler(queries, database.Pool, cfg)

	// Start background worker for hold TTL auto-release
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				released, err := queries.BulkReleaseExpiredHolds(ctx)
				if err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("⚠️  Hold expiration worker error: %v", err)
				} else if released > 0 {
					log.Printf("🧹 Auto-released %d expired seat holds", released)
				}
			}
		}
	}()

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
		// Public Auth Routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)

			// Protected /me endpoint
			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.Authenticate(cfg.JWTSecret))
				r.Get("/me", authHandler.Me)
			})
		})

		// Public/Authenticated Venue & Category Read Routes
		r.Get("/categories", venueHandler.ListCategories)
		r.Get("/venues", venueHandler.ListVenues)
		r.Get("/venues/{id}", venueHandler.GetVenue)
		r.Get("/venues/{id}/seats", venueHandler.GetVenueSeats)

		// Public Customer Event Discovery & Visual Seat Map
		r.Get("/events", eventHandler.ListPublishedEvents)
		r.Get("/events/{id}", eventHandler.GetPublicEvent)
		r.Get("/events/{id}/pricing", eventHandler.GetEventPricing)
		r.Get("/events/{id}/seats", reservationHandler.GetEventSeatMap)

		// Protected Seat Hold & Release Endpoints
		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(cfg.JWTSecret))
			r.Post("/events/{id}/hold", reservationHandler.HoldSeats)
			r.Post("/events/{id}/release", reservationHandler.ReleaseHold)
			r.Post("/bookings/checkout", bookingHandler.Checkout)
		})

		// Protected Customer Bookings Management
		r.Route("/customer", func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(cfg.JWTSecret))
			r.Use(appmiddleware.CustomerOnly)

			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				handlers.RespondSuccess(w, http.StatusOK, map[string]string{"role": "CUSTOMER"}, "Customer access verified")
			})

			r.Get("/bookings", bookingHandler.ListCustomerBookings)
			r.Get("/bookings/{id}", bookingHandler.GetBooking)
			r.Post("/bookings/{id}/cancel", bookingHandler.CancelBooking)
		})

		// Protected Admin Routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(cfg.JWTSecret))
			r.Use(appmiddleware.AdminOnly)

			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				handlers.RespondSuccess(w, http.StatusOK, map[string]string{"role": "ADMIN"}, "Admin access verified")
			})

			// Categories Admin CRUD
			r.Post("/categories", venueHandler.CreateCategory)

			// Venues Admin CRUD & Layout Configuration
			r.Post("/venues", venueHandler.CreateVenue)
			r.Put("/venues/{id}", venueHandler.UpdateVenue)
			r.Delete("/venues/{id}", venueHandler.DeleteVenue)

			// Seat Layout Endpoints
			r.Post("/venues/{id}/seats", venueHandler.CreateSeat)
			r.Post("/venues/{id}/seats/batch", venueHandler.BatchCreateSeats)
			r.Delete("/venues/{id}/seats", venueHandler.DeleteVenueSeats)
		})

		// Protected Organiser Routes
		r.Route("/organiser", func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(cfg.JWTSecret))
			r.Use(appmiddleware.OrganiserOrAdmin)

			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				handlers.RespondSuccess(w, http.StatusOK, map[string]string{"role": "ORGANISER"}, "Organiser access verified")
			})

			// Organiser Event Management
			r.Post("/events", eventHandler.CreateEvent)
			r.Get("/events", eventHandler.ListOrganiserEvents)
			r.Get("/events/{id}", eventHandler.GetOrganiserEvent)
			r.Put("/events/{id}", eventHandler.UpdateEvent)
			r.Post("/events/{id}/publish", eventHandler.PublishEvent)
			r.Post("/events/{id}/cancel", eventHandler.CancelEvent)

			// Per-Category Pricing Configuration
			r.Post("/events/{id}/pricing", eventHandler.SetEventPricing)

			// Organiser Analytics & Booking Summary
			r.Get("/events/{id}/analytics", eventHandler.GetEventAnalytics)
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
