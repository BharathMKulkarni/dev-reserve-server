package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/devreserve/server/config"
	"github.com/devreserve/server/models"
	"github.com/devreserve/server/utils"
)

// ContextKey is a type for context keys
type ContextKey string

// UserContextKey is the key for the user context
const UserContextKey ContextKey = "user"

// AuthMiddleware is middleware for authenticating requests
func AuthMiddleware(cfg config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if the Authorization header has the Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate the token
			claims, err := utils.ValidateToken(tokenString, cfg)
			if err != nil {
				http.Error(w, "Invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Create a user object from the claims
			user := models.User{
				Username: claims.Username,
				Role:     claims.Role,
			}

			// Add the user to the request context
			ctx := context.WithValue(r.Context(), UserContextKey, user)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware is middleware for restricting access to admin users
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user from the context
		userValue := r.Context().Value(UserContextKey)
		if userValue == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert the user value to a User struct
		user, ok := userValue.(models.User)
		if !ok {
			http.Error(w, "Invalid user context", http.StatusInternalServerError)
			return
		}

		// Check if the user is an admin
		if user.Role != models.RoleAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
