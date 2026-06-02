package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/user/hopper/internal/middleware"
	"github.com/user/hopper/internal/rcon"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type RconHandler struct {
	db      *sql.DB
	manager *rcon.Manager
}

func NewRconHandler(db *sql.DB, manager *rcon.Manager) *RconHandler {
	return &RconHandler{db: db, manager: manager}
}

type sessionCreds struct {
	host     string
	port     int
	password string
}

func (h *RconHandler) loadCreds(r *http.Request, id, userID string) (*sessionCreds, error) {
	var c sessionCreds
	err := h.db.QueryRowContext(r.Context(),
		`SELECT host, port, password FROM rcon_sessions WHERE id = ? AND user_id = ?`, id, userID,
	).Scan(&c.host, &c.port, &c.password)
	return &c, err
}

func (h *RconHandler) Test(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	creds, err := h.loadCreds(r, id, user.ID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	s := h.manager.GetOrCreate(id)
	if err := s.Test(creds.host, creds.port, creds.password); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "connection successful"})
}

func (h *RconHandler) Connect(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	creds, err := h.loadCreds(r, id, user.ID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	s := h.manager.GetOrCreate(id)
	if err := s.Connect(creds.host, creds.port, creds.password); err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *RconHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	var dummy string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id FROM rcon_sessions WHERE id = ? AND user_id = ?`, id, user.ID,
	).Scan(&dummy)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	if s := h.manager.Get(id); s != nil {
		s.Disconnect()
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}

func (h *RconHandler) Status(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	var dummy string
	err := h.db.QueryRowContext(r.Context(),
		`SELECT id FROM rcon_sessions WHERE id = ? AND user_id = ?`, id, user.ID,
	).Scan(&dummy)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	s := h.manager.Get(id)
	connected := s != nil && s.IsConnected()
	subscribers := 0
	if s != nil {
		subscribers = s.SubscriberCount()
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"connected":   connected,
		"subscribers": subscribers,
	})
}

func (h *RconHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUser(r)
	id := chi.URLParam(r, "id")

	creds, err := h.loadCreds(r, id, user.ID)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Time{})
	conn.SetWriteDeadline(time.Time{})

	s := h.manager.GetOrCreate(id)

	s.AttachWS(conn, user.ID, func(command, userID string) error {
		if !s.IsConnected() {
			if err := s.Connect(creds.host, creds.port, creds.password); err != nil {
				return err
			}
		}

		output, err := s.Execute(command, userID)
		if err != nil {
			return err
		}

		entryID := uuid.New().String()
		h.db.Exec(
			`INSERT INTO command_history (id, rcon_session_id, user_id, command, output, executed_at) VALUES (?, ?, ?, ?, ?, ?)`,
			entryID, id, userID, command, output, time.Now(),
		)
		return nil
	})
}
