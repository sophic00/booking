package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ticket-booking/backend/internal/auth"
)

const testSecret = "my-test-jwt-secret-key-12345678"

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := GetClaims(r.Context())
	if !ok {
		http.Error(w, "no claims", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok: " + claims.Email + " as " + claims.Role))
}

func TestAuthenticate_MissingHeader(t *testing.T) {
	handler := Authenticate(testSecret)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "UNAUTHORIZED")
}

func TestAuthenticate_MalformedHeader(t *testing.T) {
	handler := Authenticate(testSecret)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Basic 12345")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_AUTH_HEADER")
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	handler := Authenticate(testSecret)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "INVALID_TOKEN")
}

func TestAuthenticate_ValidToken(t *testing.T) {
	userID := uuid.New()
	email := "alice@example.com"
	role := "CUSTOMER"

	tokenStr, _, err := auth.GenerateToken(userID, email, role, testSecret, 1)
	require.NoError(t, err)

	handler := Authenticate(testSecret)(http.HandlerFunc(dummyHandler))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok: alice@example.com as CUSTOMER", rr.Body.String())
}

func TestRBAC_AdminOnly(t *testing.T) {
	adminID := uuid.New()
	adminToken, _, _ := auth.GenerateToken(adminID, "admin@example.com", "ADMIN", testSecret, 1)

	customerID := uuid.New()
	customerToken, _, _ := auth.GenerateToken(customerID, "customer@example.com", "CUSTOMER", testSecret, 1)

	handler := Authenticate(testSecret)(AdminOnly(http.HandlerFunc(dummyHandler)))

	// 1. Admin request -> should succeed (200)
	reqAdmin := httptest.NewRequest(http.MethodGet, "/admin", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	rrAdmin := httptest.NewRecorder()
	handler.ServeHTTP(rrAdmin, reqAdmin)
	assert.Equal(t, http.StatusOK, rrAdmin.Code)

	// 2. Customer request -> should be forbidden (403)
	reqCustomer := httptest.NewRequest(http.MethodGet, "/admin", nil)
	reqCustomer.Header.Set("Authorization", "Bearer "+customerToken)
	rrCustomer := httptest.NewRecorder()
	handler.ServeHTTP(rrCustomer, reqCustomer)
	assert.Equal(t, http.StatusForbidden, rrCustomer.Code)
	assert.Contains(t, rrCustomer.Body.String(), "FORBIDDEN")
}

func TestRBAC_OrganiserOnly(t *testing.T) {
	organiserID := uuid.New()
	organiserToken, _, _ := auth.GenerateToken(organiserID, "org@example.com", "ORGANISER", testSecret, 1)

	customerID := uuid.New()
	customerToken, _, _ := auth.GenerateToken(customerID, "customer@example.com", "CUSTOMER", testSecret, 1)

	handler := Authenticate(testSecret)(OrganiserOnly(http.HandlerFunc(dummyHandler)))

	// 1. Organiser -> 200
	reqOrg := httptest.NewRequest(http.MethodGet, "/organiser", nil)
	reqOrg.Header.Set("Authorization", "Bearer "+organiserToken)
	rrOrg := httptest.NewRecorder()
	handler.ServeHTTP(rrOrg, reqOrg)
	assert.Equal(t, http.StatusOK, rrOrg.Code)

	// 2. Customer -> 403
	reqCust := httptest.NewRequest(http.MethodGet, "/organiser", nil)
	reqCust.Header.Set("Authorization", "Bearer "+customerToken)
	rrCust := httptest.NewRecorder()
	handler.ServeHTTP(rrCust, reqCust)
	assert.Equal(t, http.StatusForbidden, rrCust.Code)
}

func TestRBAC_CustomerOnly(t *testing.T) {
	customerID := uuid.New()
	customerToken, _, _ := auth.GenerateToken(customerID, "cust@example.com", "CUSTOMER", testSecret, 1)

	adminID := uuid.New()
	adminToken, _, _ := auth.GenerateToken(adminID, "admin@example.com", "ADMIN", testSecret, 1)

	handler := Authenticate(testSecret)(CustomerOnly(http.HandlerFunc(dummyHandler)))

	// 1. Customer -> 200
	reqCust := httptest.NewRequest(http.MethodGet, "/customer", nil)
	reqCust.Header.Set("Authorization", "Bearer "+customerToken)
	rrCust := httptest.NewRecorder()
	handler.ServeHTTP(rrCust, reqCust)
	assert.Equal(t, http.StatusOK, rrCust.Code)

	// 2. Admin -> 403 (CustomerOnly strictly requires CUSTOMER)
	reqAdmin := httptest.NewRequest(http.MethodGet, "/customer", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	rrAdmin := httptest.NewRecorder()
	handler.ServeHTTP(rrAdmin, reqAdmin)
	assert.Equal(t, http.StatusForbidden, rrAdmin.Code)
}

func TestRBAC_OrganiserOrAdmin(t *testing.T) {
	handler := Authenticate(testSecret)(OrganiserOrAdmin(http.HandlerFunc(dummyHandler)))

	adminID := uuid.New()
	adminToken, _, _ := auth.GenerateToken(adminID, "admin@example.com", "ADMIN", testSecret, 1)

	orgID := uuid.New()
	orgToken, _, _ := auth.GenerateToken(orgID, "org@example.com", "ORGANISER", testSecret, 1)

	custID := uuid.New()
	custToken, _, _ := auth.GenerateToken(custID, "cust@example.com", "CUSTOMER", testSecret, 1)

	// Admin -> 200
	req := httptest.NewRequest(http.MethodGet, "/manage", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Organiser -> 200
	req = httptest.NewRequest(http.MethodGet, "/manage", nil)
	req.Header.Set("Authorization", "Bearer "+orgToken)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// Customer -> 403
	req = httptest.NewRequest(http.MethodGet, "/manage", nil)
	req.Header.Set("Authorization", "Bearer "+custToken)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestContextHelpers_NoClaims(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := req.Context()

	_, err := GetUserID(ctx)
	assert.ErrorIs(t, err, ErrNoClaimsInContext)

	_, err = GetUserRole(ctx)
	assert.ErrorIs(t, err, ErrNoClaimsInContext)

	_, err = GetUserEmail(ctx)
	assert.ErrorIs(t, err, ErrNoClaimsInContext)
}
