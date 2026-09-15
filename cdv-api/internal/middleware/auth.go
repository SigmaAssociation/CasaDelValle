package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const RoleAdmin = 1

type contextKey string

const (
	userIDKey    contextKey = "auth_user_id"
	userRoleKey  contextKey = "auth_user_role"
	userEmailKey contextKey = "auth_user_email"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		if isPublicRoute(r) {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeAuthError(w, http.StatusUnauthorized, "Token de autenticación no proporcionado")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeAuthError(w, http.StatusUnauthorized, "Formato de token inválido. Use: Bearer <token>")
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			writeAuthError(w, http.StatusInternalServerError, "Error interno del servidor")
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			writeAuthError(w, http.StatusUnauthorized, "Token inválido o expirado")
			return
		}

		userID := claimToUint(claims["sub"])
		role := claimToUint(claims["role"])
		if userID == 0 {
			writeAuthError(w, http.StatusUnauthorized, "Token inválido o expirado")
			return
		}

		email, _ := claims["email"].(string)

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, userRoleKey, role)
		ctx = context.WithValue(ctx, userEmailKey, email)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicRoute(r *http.Request) bool {
	if r.Method != http.MethodPost {
		return false
	}

	switch r.URL.Path {
	case "/cdv-api/login", "/cdv-api/users":
		return true
	default:
		return false
	}
}

func UserIDFromContext(ctx context.Context) (uint, bool) {
	id, ok := ctx.Value(userIDKey).(uint)
	return id, ok
}

func RoleFromContext(ctx context.Context) (uint, bool) {
	role, ok := ctx.Value(userRoleKey).(uint)
	return role, ok
}

func IsAdmin(r *http.Request) bool {
	role, ok := RoleFromContext(r.Context())
	return ok && role == RoleAdmin
}

func AuthorizeSelfOrAdmin(r *http.Request, targetUserID uint) bool {
	authID, ok := UserIDFromContext(r.Context())
	if !ok {
		return false
	}
	if authID == targetUserID {
		return true
	}
	return IsAdmin(r)
}

func claimToUint(v any) uint {
	switch n := v.(type) {
	case float64:
		if n <= 0 {
			return 0
		}
		return uint(n)
	case float32:
		if n <= 0 {
			return 0
		}
		return uint(n)
	case int:
		if n <= 0 {
			return 0
		}
		return uint(n)
	case int64:
		if n <= 0 {
			return 0
		}
		return uint(n)
	case uint:
		return n
	case uint64:
		return uint(n)
	default:
		return 0
	}
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}
