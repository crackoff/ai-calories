package handler

import (
	"ai-calories/internal/model"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	jwtpkg "ai-calories/pkg/jwt"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "missing or invalid authorization header"})
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims, err := jwtpkg.ValidateAccessToken(tokenStr, jwtSecret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "invalid or expired token"})
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getUserID(r *http.Request) uint {
	return r.Context().Value(UserIDKey).(uint)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
