package rcon

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	gorcon "github.com/gorcon/rcon"
	"github.com/gorilla/websocket"
)

type WsMessage struct {
	Type      string    `json:"type"`
	ID        string    `json:"id,omitempty"`
	Command   string    `json:"command,omitempty"`
	Output    string    `json:"output,omitempty"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	UserID    string    `json:"userId,omitempty"`
}

type cmdRequest struct {
	command  string
	userID   string
	resultCh chan cmdResult
}

type cmdResult struct {
	output string
	err    error
}

type subscriber struct {
	writeCh chan []byte
}

type Session struct {
	id        string
	conn      *gorcon.Conn
	connMu    sync.Mutex
	connected bool
	cmdCh     chan cmdRequest
	subs      map[string]*subscriber
	subsMu    sync.RWMutex
	stopCh    chan struct{}
	stopOnce  sync.Once
}

type Manager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{sessions: make(map[string]*Session)}
}

func (m *Manager) GetOrCreate(id string) *Session {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		return s
	}
	s := &Session{
		id:     id,
		cmdCh:  make(chan cmdRequest, 32),
		subs:   make(map[string]*subscriber),
		stopCh: make(chan struct{}),
	}
	m.sessions[id] = s
	return s
}

func (m *Manager) Get(id string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[id]
}

func (m *Manager) Remove(id string) {
	m.mu.Lock()
	s, ok := m.sessions[id]
	if ok {
		delete(m.sessions, id)
	}
	m.mu.Unlock()
	if ok {
		s.shutdown()
	}
}

func (s *Session) Connect(host string, port int, password string) error {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if s.connected {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := gorcon.Dial(addr, password)
	if err != nil {
		return err
	}
	s.conn = conn
	s.connected = true

	go s.runCommandProcessor()
	return nil
}

func (s *Session) Test(host string, port int, password string) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := gorcon.Dial(addr, password)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (s *Session) Disconnect() {
	s.connMu.Lock()
	defer s.connMu.Unlock()

	if !s.connected {
		return
	}
	s.conn.Close()
	s.conn = nil
	s.connected = false

	s.broadcast(WsMessage{Type: "disconnected"})
}

func (s *Session) IsConnected() bool {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	return s.connected
}

func (s *Session) SubscriberCount() int {
	s.subsMu.RLock()
	defer s.subsMu.RUnlock()
	return len(s.subs)
}

func (s *Session) Execute(command, userID string) (string, error) {
	resultCh := make(chan cmdResult, 1)
	select {
	case s.cmdCh <- cmdRequest{command: command, userID: userID, resultCh: resultCh}:
	case <-s.stopCh:
		return "", fmt.Errorf("session stopped")
	}
	r := <-resultCh
	return r.output, r.err
}

func (s *Session) runCommandProcessor() {
	for {
		select {
		case req := <-s.cmdCh:
			s.connMu.Lock()
			if !s.connected || s.conn == nil {
				s.connMu.Unlock()
				req.resultCh <- cmdResult{err: fmt.Errorf("not connected")}
				continue
			}
			output, err := s.conn.Execute(req.command)
			s.connMu.Unlock()

			req.resultCh <- cmdResult{output: output, err: err}

			if err == nil {
				s.broadcast(WsMessage{
					Type:      "output",
					ID:        uuid.New().String(),
					Command:   req.command,
					Output:    output,
					Timestamp: time.Now(),
					UserID:    req.userID,
				})
			} else {
				s.broadcast(WsMessage{Type: "error", Message: err.Error()})
				s.connMu.Lock()
				s.connected = false
				s.connMu.Unlock()
			}
		case <-s.stopCh:
			return
		}
	}
}

func (s *Session) broadcast(msg WsMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	s.subsMu.RLock()
	defer s.subsMu.RUnlock()
	for _, sub := range s.subs {
		select {
		case sub.writeCh <- data:
		default:
		}
	}
}

func (s *Session) AttachWS(conn *websocket.Conn, userID string, onCommand func(command, userID string) error) {
	subID := uuid.New().String()
	sub := &subscriber{writeCh: make(chan []byte, 64)}

	s.subsMu.Lock()
	s.subs[subID] = sub
	s.subsMu.Unlock()

	defer func() {
		s.subsMu.Lock()
		delete(s.subs, subID)
		s.subsMu.Unlock()
		close(sub.writeCh)
	}()

	go func() {
		for data := range sub.writeCh {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				conn.Close()
				return
			}
		}
	}()

	statusType := "disconnected"
	if s.IsConnected() {
		statusType = "connected"
	}
	if data, err := json.Marshal(WsMessage{Type: statusType}); err == nil {
		select {
		case sub.writeCh <- data:
		default:
		}
	}

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			return
		}
		var msg struct {
			Type string `json:"type"`
			Data string `json:"data"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		if msg.Type == "command" && msg.Data != "" {
			if err := onCommand(msg.Data, userID); err != nil {
				if data, e := json.Marshal(WsMessage{Type: "error", Message: err.Error()}); e == nil {
					select {
					case sub.writeCh <- data:
					default:
					}
				}
			}
		}
	}
}

func (s *Session) shutdown() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		s.Disconnect()
	})
}
