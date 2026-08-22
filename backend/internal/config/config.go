package config

import (
	"log"
	"os"
	"strconv"
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
	CORSAllowedOrigins string
	Environment        string

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
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
		Environment:        getEnv("APP_ENV", "development"),

		SMTPServer:   getEnv("SMTP_SERVER", "localhost"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 1025),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "no-reply@ticketbooking.local"),
		EmailMock:    getEnvAsBool("EMAIL_MOCK", true),
	}

	return cfg
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
