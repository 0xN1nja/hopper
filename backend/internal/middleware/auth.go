package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/user/hopper/internal/models"
)

type contextKey string

const UserKey contextKey = "user"

func Auth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := tokenFromRequest(r)
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			var user models.User
			var expiresAt time.Time
			err := db.QueryRowContext(r.Context(), `
				SELECT u.id, u.username, s.expires_at
				FROM auth_sessions s
				JOIN users u ON u.id = s.user_id
				WHERE s.token = ?
			`, token).Scan(&user.ID, &user.Username, &expiresAt)

			if err == sql.ErrNoRows || time.Now().After(expiresAt) {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			if err != nil {
				http.Error(w, `{"error":"server error"}`, http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), UserKey, &user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUser(r *http.Request) *models.User {
	user, _ := r.Context().Value(UserKey).(*models.User)
	return user
}

func tokenFromRequest(r *http.Request) string {
	if cookie, err := r.Cookie("session"); err == nil {
		return cookie.Value
	}
	if t := r.URL.Query().Get("token"); t != "" {
		return t
	}
	return ""
}
