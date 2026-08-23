package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	JWTExpirationHours int
	SeatHoldTTL        time.Duration
	WaitlistOfferTTL   time.Duration
	CORSAllowedOrigins []string
	Environment        string
	FrontendBaseURL    string

	// Email config
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	EmailMock    bool
}

func Load() *Config {
	// Try loading .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found or could not be loaded, using environment variables")
	}

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/ticket_booking?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret-jwt-key-for-ticket-booking-dev"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 72),
		SeatHoldTTL:        time.Duration(getEnvAsInt("SEAT_HOLD_TTL_MINUTES", 10)) * time.Minute,
		WaitlistOfferTTL:   time.Duration(getEnvAsInt("WAITLIST_OFFER_TTL_MINUTES", 10)) * time.Minute,
		CORSAllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8080"}, ","),
		Environment:        getEnv("APP_ENV", "development"),
		FrontendBaseURL:    getEnv("FRONTEND_BASE_URL", "http://localhost:3000"),

		SMTPServer:   getEnv("SMTP_SERVER", "localhost"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 1025),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "no-reply@ticketbooking.local"),
		EmailMock:    getEnvAsBool("EMAIL_MOCK", true),
	}

	return cfg
}

func getEnvAsSlice(key string, fallback []string, sep string) []string {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return fallback
	}
	parts := strings.Split(val, sep)
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strVal := getEnv(key, "")
	if strVal == "" {
		return fallback
	}
	val, err := strconv.Atoi(strVal)
	if err != nil {
		return fallback
	}
	return val
}

func getEnvAsBool(key string, fallback bool) bool {
	strVal := getEnv(key, "")
	if strVal == "" {
		return fallback
	}
	val, err := strconv.ParseBool(strVal)
	if err != nil {
		return fallback
	}
	return val
}
