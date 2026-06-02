package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type RconSession struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	CreatedAt time.Time `json:"createdAt"`
}

type CommandEntry struct {
	ID            string    `json:"id"`
	RconSessionID string    `json:"rconSessionId"`
	UserID        string    `json:"userId"`
	Command       string    `json:"command"`
	Output        string    `json:"output"`
	ExecutedAt    time.Time `json:"executedAt"`
}
