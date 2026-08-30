package http

import (
	"context"
	"net/http"
	"strings"

	"ne/internal/auth"
	"ne/internal/domain"
)

type contextKey string

const userContextKey contextKey = "user_claims"

// AuthMiddleware extracts and validates JWT tokens from Authorization headers or stream query params.
func AuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := jwtManager.VerifyAccessToken(tokenStr)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves authenticated user claims from the request context.
func GetUserFromContext(ctx context.Context) *auth.UserClaims {
	if val := ctx.Value(userContextKey); val != nil {
		if claims, ok := val.(*auth.UserClaims); ok {
			return claims
		}
	}
	return nil
}

// GetAuthClaims is an alias for GetUserFromContext.
func GetAuthClaims(ctx context.Context) *auth.UserClaims {
	return GetUserFromContext(ctx)
}

// RequireAuth blocks unauthenticated requests with HTTP 401.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin blocks non-admin requests with HTTP 403.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user == nil {
			RespondError(w, r, domain.ErrUnauthorized("Authentication required"))
			return
		}
		if !user.IsAdmin {
			RespondError(w, r, domain.ErrForbidden("Administrator privileges required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	// 1. Authorization header: Bearer <token>
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. Query param (for audio/image streaming): ?token=... or ?jwt=...
	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}
	if token := r.URL.Query().Get("jwt"); token != "" {
		return token
	}

	return ""
}
