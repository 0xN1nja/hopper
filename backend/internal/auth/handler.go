package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/user/hopper/internal/middleware"
	"github.com/user/hopper/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password required")
		return
	}

	var user models.User
	var hash string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id, username, password_hash FROM users WHERE username = ?`, req.Username,
	).Scan(&user.ID, &user.Username, &hash)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := uuid.New().String()
	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	_, err = h.db.ExecContext(r.Context(),
		`INSERT INTO auth_sessions (id, user_id, token, expires_at) VALUES (?, ?, ?, ?)`,
		sessionID, user.ID, token, expiresAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("session"); err == nil {
		h.db.ExecContext(r.Context(), `DELETE FROM auth_sessions WHERE token = ?`, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func SeedUser(db *sql.DB, username, password string) error {
	if password == "" {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	id := uuid.New().String()
	_, err = db.Exec(`
		INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)
		ON CONFLICT(username) DO UPDATE SET password_hash = excluded.password_hash
	`, id, username, string(hash))
	return err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
