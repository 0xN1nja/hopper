package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/user/hopper/internal/middleware"
	"github.com/user/hopper/internal/models"
	"github.com/user/hopper/internal/rcon"
)

type SessionsHandler struct {
	db      *sql.DB
	manager *rcon.Manager
}

func NewSessionsHandler(db *sql.DB, manager *rcon.Manager) *SessionsHandler {
	return &SessionsHandler{db: db, manager: manager}
}

func (h *SessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, name, icon, host, port, created_at FROM rcon_sessions WHERE user_id = ? ORDER BY created_at DESC`,
		user.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	defer rows.Close()

	sessions := []map[string]any{}
	for rows.Next() {
		var s models.RconSession
		if err := rows.Scan(&s.ID, &s.Name, &s.Icon, &s.Host, &s.Port, &s.CreatedAt); err != nil {
			continue
		}
		ms := h.manager.Get(s.ID)
		connected := ms != nil && ms.IsConnected()
		subscribers := 0
		if ms != nil {
			subscribers = ms.SubscriberCount()
		}
		sessions = append(sessions, map[string]any{
			"id":          s.ID,
			"name":        s.Name,
			"icon":        s.Icon,
			"host":        s.Host,
			"port":        s.Port,
			"createdAt":   s.CreatedAt,
			"connected":   connected,
			"subscribers": subscribers,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (h *SessionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	var req struct {
		Name     string `json:"name"`
		Icon     string `json:"icon"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" || req.Host == "" || req.Port == 0 {
		writeError(w, http.StatusBadRequest, "name, host, and port are required")
		return
	}

	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.ExecContext(r.Context(),
		`INSERT INTO rcon_sessions (id, user_id, name, icon, host, port, password, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, user.ID, req.Name, req.Icon, req.Host, req.Port, req.Password, now,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":        id,
		"name":      req.Name,
		"icon":      req.Icon,
		"host":      req.Host,
		"port":      req.Port,
		"createdAt": now,
		"connected": false,
	})
}

func (h *SessionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	var existing struct {
		password string
	}
	err := h.db.QueryRowContext(r.Context(),
		`SELECT password FROM rcon_sessions WHERE id = ? AND user_id = ?`, id, user.ID,
	).Scan(&existing.password)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	var req struct {
		Name     string `json:"name"`
		Icon     string `json:"icon"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	password := existing.password
	if req.Password != "" {
		password = req.Password
	}

	_, err = h.db.ExecContext(r.Context(),
		`UPDATE rcon_sessions SET name = ?, icon = ?, host = ?, port = ?, password = ? WHERE id = ? AND user_id = ?`,
		req.Name, req.Icon, req.Host, req.Port, password, id, user.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *SessionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	res, err := h.db.ExecContext(r.Context(),
		`DELETE FROM rcon_sessions WHERE id = ? AND user_id = ?`, id, user.ID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	h.manager.Remove(id)
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *SessionsHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	if err := h.assertOwns(r, id, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	limit := 200
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}

	rows, err := h.db.QueryContext(r.Context(),
		`SELECT id, command, output, user_id, executed_at FROM command_history WHERE rcon_session_id = ? ORDER BY executed_at ASC LIMIT ?`,
		id, limit,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	defer rows.Close()

	entries := []models.CommandEntry{}
	for rows.Next() {
		var e models.CommandEntry
		if err := rows.Scan(&e.ID, &e.Command, &e.Output, &e.UserID, &e.ExecutedAt); err != nil {
			continue
		}
		e.RconSessionID = id
		entries = append(entries, e)
	}
	writeJSON(w, http.StatusOK, map[string]any{"history": entries})
}

func (h *SessionsHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	if err := h.assertOwns(r, id, user.ID); err != nil {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	_, err := h.db.ExecContext(r.Context(),
		`DELETE FROM command_history WHERE rcon_session_id = ?`, id,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (h *SessionsHandler) assertOwns(r *http.Request, sessionID, userID string) error {
	var dummy string
	return h.db.QueryRowContext(r.Context(),
		`SELECT id FROM rcon_sessions WHERE id = ? AND user_id = ?`, sessionID, userID,
	).Scan(&dummy)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
