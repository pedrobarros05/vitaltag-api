package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/pedrobarros/vitaltag-api/internal/service"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// JWT extracts and validates the Bearer token, injecting userID into the context.
func JWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			respondError(w, http.StatusUnauthorized, "missing or malformed token")
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := service.ValidateJWT(tokenStr)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserIDFromContext extracts the authenticated user's ID.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

func respondError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
