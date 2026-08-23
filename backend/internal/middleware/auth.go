package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"ticket-booking/backend/internal/auth"
)

type contextKey string

const claimsContextKey contextKey = "jwt_claims"

var (
	ErrNoClaimsInContext = errors.New("no auth claims found in request context")
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error:  message,
		Code:   code,
		Status: status,
	})
}

// Authenticate verifies the JWT token in Authorization header and adds claims to context.
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeJSONError(w, http.StatusUnauthorized, "INVALID_AUTH_HEADER", "authorization header format must be Bearer <token>")
				return
			}

			tokenStr := strings.TrimSpace(parts[1])
			if tokenStr == "" {
				writeJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "bearer token cannot be empty")
				return
			}

			claims, err := auth.ValidateToken(tokenStr, jwtSecret)
			if err != nil {
				if errors.Is(err, auth.ErrExpiredToken) {
					writeJSONError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "authentication token has expired")
					return
				}
				writeJSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "invalid authentication token")
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetClaims retrieves auth claims from context.
func GetClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*auth.Claims)
	return claims, ok
}

// GetUserID extracts the user ID UUID from context.
func GetUserID(ctx context.Context) (uuid.UUID, error) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return uuid.Nil, ErrNoClaimsInContext
	}
	return claims.UserID, nil
}

// GetUserRole extracts the user role from context.
func GetUserRole(ctx context.Context) (string, error) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", ErrNoClaimsInContext
	}
	return claims.Role, nil
}

// GetUserEmail extracts the user email from context.
func GetUserEmail(ctx context.Context) (string, error) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", ErrNoClaimsInContext
	}
	return claims.Email, nil
}

// RequireRole ensures the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaims(r.Context())
			if !ok || claims == nil {
				writeJSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required")
				return
			}

			userRole := strings.ToUpper(strings.TrimSpace(claims.Role))
			isAllowed := false
			for _, role := range allowedRoles {
				if strings.EqualFold(userRole, strings.TrimSpace(role)) {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				writeJSONError(w, http.StatusForbidden, "FORBIDDEN", "access denied: insufficient permissions for role "+userRole)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnly restricts route to ADMIN role.
func AdminOnly(next http.Handler) http.Handler {
	return RequireRole("ADMIN")(next)
}

// OrganiserOnly restricts route to ORGANISER role.
func OrganiserOnly(next http.Handler) http.Handler {
	return RequireRole("ORGANISER")(next)
}

// CustomerOnly restricts route to CUSTOMER role.
func CustomerOnly(next http.Handler) http.Handler {
	return RequireRole("CUSTOMER")(next)
}

// OrganiserOrAdmin allows either ORGANISER or ADMIN role.
func OrganiserOrAdmin(next http.Handler) http.Handler {
	return RequireRole("ORGANISER", "ADMIN")(next)
}
