package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordHashing(t *testing.T) {
	rawPassword := "SecurePassword123!"

	hash, err := HashPassword(rawPassword)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, rawPassword, hash)

	// Valid password check
	assert.True(t, CheckPasswordHash(rawPassword, hash))

	// Invalid password check
	assert.False(t, CheckPasswordHash("WrongPassword456!", hash))
}

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-123456"
	userID := uuid.New()
	email := "customer@example.com"
	role := "CUSTOMER"

	// 1. Generate token
	tokenStr, expiresAt, err := GenerateToken(userID, email, role, secret, 2)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
	assert.True(t, expiresAt.After(time.Now()))

	// 2. Validate token
	claims, err := ValidateToken(tokenStr, secret)
	require.NoError(t, err)
	require.NotNil(t, claims)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, "ticket-booking-api", claims.Issuer)
	assert.Equal(t, userID.String(), claims.Subject)
}

func TestJWTValidation_WrongSecret(t *testing.T) {
	secret := "correct-secret"
	wrongSecret := "incorrect-secret"
	userID := uuid.New()

	tokenStr, _, err := GenerateToken(userID, "user@example.com", "ADMIN", secret, 1)
	require.NoError(t, err)

	claims, err := ValidateToken(tokenStr, wrongSecret)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTValidation_ExpiredToken(t *testing.T) {
	secret := "secret-for-expiration"
	userID := uuid.New()

	// Create an expired token manually
	claims := &Claims{
		UserID: userID,
		Email:  "expired@example.com",
		Role:   "ORGANISER",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // 1 hour in the past
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Subject:   userID.String(),
			Issuer:    "ticket-booking-api",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = ValidateToken(tokenStr, secret)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestJWTValidation_MalformedToken(t *testing.T) {
	secret := "my-secret"
	_, err := ValidateToken("invalid.malformed.jwt.token", secret)
	assert.Error(t, err)
}
