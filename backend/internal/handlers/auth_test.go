package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticket-booking/backend/internal/auth"
	"ticket-booking/backend/internal/config"
	"ticket-booking/backend/internal/db/generated"
	"ticket-booking/backend/internal/middleware"
	"ticket-booking/backend/internal/utils"
)

// mockQuerier implements generated.Querier for testing
type mockQuerier struct {
	generated.Querier
	createUserFunc     func(ctx context.Context, arg generated.CreateUserParams) (generated.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (generated.User, error)
	getUserByIDFunc    func(ctx context.Context, id pgtype.UUID) (generated.User, error)
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg generated.CreateUserParams) (generated.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, arg)
	}
	return generated.User{}, errorsNew("unimplemented")
}

func (m *mockQuerier) GetUserByEmail(ctx context.Context, email string) (generated.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return generated.User{}, pgx.ErrNoRows
}

func (m *mockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (generated.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return generated.User{}, pgx.ErrNoRows
}

func errorsNew(msg string) error {
	return &pgconn.PgError{Message: msg}
}

func testConfig() *config.Config {
	return &config.Config{
		JWTSecret:          "test-jwt-secret-key",
		JWTExpirationHours: 24,
	}
}

func TestRegister_Success(t *testing.T) {
	cfg := testConfig()
	mock := &mockQuerier{
		createUserFunc: func(ctx context.Context, arg generated.CreateUserParams) (generated.User, error) {
			assert.Equal(t, "user@example.com", arg.Email)
			assert.Equal(t, "John Doe", arg.FullName)
			assert.Equal(t, generated.UserRoleCUSTOMER, arg.Role)
			assert.True(t, auth.CheckPasswordHash("Password123!", arg.PasswordHash))

			return generated.User{
				ID:           utils.UUIDToPgtype(uuid.New()),
				Email:        arg.Email,
				FullName:     arg.FullName,
				Role:         arg.Role,
				PasswordHash: arg.PasswordHash,
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewAuthHandler(mock, cfg)

	body := RegisterRequest{
		Email:    "user@example.com",
		Password: "Password123!",
		FullName: "John Doe",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp struct {
		Success bool         `json:"success"`
		Data    AuthResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Data.Token)
	assert.Equal(t, "user@example.com", resp.Data.User.Email)
	assert.Equal(t, "CUSTOMER", resp.Data.User.Role)
}

func TestRegister_InvalidEmail(t *testing.T) {
	handler := NewAuthHandler(&mockQuerier{}, testConfig())

	body := RegisterRequest{
		Email:    "not-an-email",
		Password: "Password123!",
		FullName: "John Doe",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_EMAIL")
}

func TestRegister_ShortPassword(t *testing.T) {
	handler := NewAuthHandler(&mockQuerier{}, testConfig())

	body := RegisterRequest{
		Email:    "user@example.com",
		Password: "short",
		FullName: "John Doe",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_PASSWORD")
}

func TestRegister_DuplicateEmail(t *testing.T) {
	mock := &mockQuerier{
		createUserFunc: func(ctx context.Context, arg generated.CreateUserParams) (generated.User, error) {
			return generated.User{}, &pgconn.PgError{Code: pgerrcode.UniqueViolation}
		},
	}

	handler := NewAuthHandler(mock, testConfig())

	body := RegisterRequest{
		Email:    "duplicate@example.com",
		Password: "Password123!",
		FullName: "Duplicate User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Register(rr, req)

	assert.Equal(t, http.StatusConflict, rr.Code)
	assert.Contains(t, rr.Body.String(), "EMAIL_EXISTS")
}

func TestLogin_Success(t *testing.T) {
	cfg := testConfig()
	password := "SecretPass123!"
	hashed, _ := auth.HashPassword(password)
	userID := uuid.New()

	mock := &mockQuerier{
		getUserByEmailFunc: func(ctx context.Context, email string) (generated.User, error) {
			return generated.User{
				ID:           utils.UUIDToPgtype(userID),
				Email:        email,
				FullName:     "Organiser One",
				Role:         generated.UserRoleORGANISER,
				PasswordHash: hashed,
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewAuthHandler(mock, cfg)

	body := LoginRequest{
		Email:    "org@example.com",
		Password: password,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool         `json:"success"`
		Data    AuthResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Data.Token)
	assert.Equal(t, "ORGANISER", resp.Data.User.Role)
	assert.Equal(t, userID.String(), resp.Data.User.ID)
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := auth.HashPassword("CorrectPassword123!")

	mock := &mockQuerier{
		getUserByEmailFunc: func(ctx context.Context, email string) (generated.User, error) {
			return generated.User{
				Email:        email,
				PasswordHash: hashed,
			}, nil
		},
	}

	handler := NewAuthHandler(mock, testConfig())

	body := LoginRequest{
		Email:    "user@example.com",
		Password: "IncorrectPassword!",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_CREDENTIALS")
}

func TestLogin_UserNotFound(t *testing.T) {
	mock := &mockQuerier{
		getUserByEmailFunc: func(ctx context.Context, email string) (generated.User, error) {
			return generated.User{}, pgx.ErrNoRows
		},
	}

	handler := NewAuthHandler(mock, testConfig())

	body := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "SomePassword123!",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
	rr := httptest.NewRecorder()

	handler.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_CREDENTIALS")
}

func TestMe_Success(t *testing.T) {
	cfg := testConfig()
	userID := uuid.New()

	mock := &mockQuerier{
		getUserByIDFunc: func(ctx context.Context, id pgtype.UUID) (generated.User, error) {
			return generated.User{
				ID:        id,
				Email:     "me@example.com",
				FullName:  "Alice Customer",
				Role:      generated.UserRoleCUSTOMER,
				CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	handler := NewAuthHandler(mock, cfg)

	token, _, _ := auth.GenerateToken(userID, "me@example.com", "CUSTOMER", cfg.JWTSecret, 1)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Wrap with Authenticate middleware
	authMiddleware := middleware.Authenticate(cfg.JWTSecret)
	authMiddleware(http.HandlerFunc(handler.Me)).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp struct {
		Success bool         `json:"success"`
		Data    UserResponse `json:"data"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "me@example.com", resp.Data.Email)
	assert.Equal(t, "Alice Customer", resp.Data.FullName)
}
