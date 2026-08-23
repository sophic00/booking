package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"ticket-booking/backend/internal/auth"
	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

type AuthHandler struct {
	queries generated.Querier
	cfg     *config.Config
}

func NewAuthHandler(queries generated.Querier, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		queries: queries,
		cfg:     cfg,
	}
}

type RegisterRequest struct {
	Email    string  `json:"email"`
	Password string  `json:"password"`
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone,omitempty"`
	Role     *string `json:"role,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Phone     *string   `json:"phone,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func formatUserResponse(u generated.User) UserResponse {
	var phone *string
	if u.Phone.Valid && u.Phone.String != "" {
		phone = &u.Phone.String
	}

	return UserResponse{
		ID:        utils.PgtypeToUUID(u.ID).String(),
		Email:     u.Email,
		FullName:  u.FullName,
		Phone:     phone,
		Role:      string(u.Role),
		CreatedAt: utils.PgtypeToTime(u.CreatedAt),
	}
}

// Register handles new user registration.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	// Validate email
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if _, err := mail.ParseAddress(email); err != nil || email == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_EMAIL", "please provide a valid email address")
		return
	}

	// Validate password
	if len(req.Password) < 8 {
		RespondError(w, http.StatusBadRequest, "INVALID_PASSWORD", "password must be at least 8 characters long")
		return
	}

	// Validate full name
	fullName := strings.TrimSpace(req.FullName)
	if fullName == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_NAME", "full name cannot be empty")
		return
	}

	// Validate role (default: CUSTOMER)
	role := generated.UserRoleCUSTOMER
	if req.Role != nil && strings.TrimSpace(*req.Role) != "" {
		switch strings.ToUpper(strings.TrimSpace(*req.Role)) {
		case "ADMIN":
			role = generated.UserRoleADMIN
		case "ORGANISER":
			role = generated.UserRoleORGANISER
		case "CUSTOMER":
			role = generated.UserRoleCUSTOMER
		default:
			RespondError(w, http.StatusBadRequest, "INVALID_ROLE", "invalid role: must be CUSTOMER, ORGANISER, or ADMIN")
			return
		}
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to process credentials")
		return
	}

	phoneText := pgtype.Text{Valid: false}
	if req.Phone != nil && strings.TrimSpace(*req.Phone) != "" {
		phoneText = pgtype.Text{String: strings.TrimSpace(*req.Phone), Valid: true}
	}

	user, err := h.queries.CreateUser(r.Context(), generated.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		Phone:        phoneText,
		Role:         role,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "user with this email already exists")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to create user account")
		return
	}

	userID := utils.PgtypeToUUID(user.ID)
	token, expiresAt, err := auth.GenerateToken(userID, user.Email, string(user.Role), h.cfg.JWTSecret, h.cfg.JWTExpirationHours)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to generate authentication token")
		return
	}

	RespondSuccess(w, http.StatusCreated, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      formatUserResponse(user),
	}, "Account registered successfully")
}

// Login authenticates a user and returns a JWT.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "malformed JSON body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		RespondError(w, http.StatusBadRequest, "INVALID_CREDENTIALS", "email and password are required")
		return
	}

	user, err := h.queries.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to retrieve user")
		return
	}

	if !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password")
		return
	}

	userID := utils.PgtypeToUUID(user.ID)
	token, expiresAt, err := auth.GenerateToken(userID, user.Email, string(user.Role), h.cfg.JWTSecret, h.cfg.JWTExpirationHours)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to generate authentication token")
		return
	}

	RespondSuccess(w, http.StatusOK, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      formatUserResponse(user),
	}, "Login successful")
}

// Me retrieves the current logged-in user profile.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), utils.UUIDToPgtype(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(w, http.StatusNotFound, "USER_NOT_FOUND", "user profile not found")
			return
		}
		RespondError(w, http.StatusInternalServerError, "DB_ERROR", "failed to fetch user profile")
		return
	}

	RespondSuccess(w, http.StatusOK, formatUserResponse(user), "User profile retrieved")
}
