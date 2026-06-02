package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func New(path string) (*sql.DB, error) {
	dsn := path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	database.SetMaxOpenConns(1)
	if err := database.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(database); err != nil {
		return nil, err
	}
	return database, nil
}

func migrateUsersTable(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(users)`)
	if err != nil {
		return err
	}

	cols := map[string]bool{}
	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var dflt interface{}
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			rows.Close()
			return err
		}
		cols[name] = true
	}
	if err := rows.Close(); err != nil {
		return err
	}

	if cols["username"] && !cols["name"] {
		return nil
	}
	if !cols["username"] && !cols["email"] && !cols["name"] {
		return nil
	}

	loginCol := "email"
	if cols["username"] {
		loginCol = "username"
	} else if !cols["email"] && cols["name"] {
		loginCol = "name"
	}

	steps := []string{
		`PRAGMA foreign_keys = OFF`,
		`DROP TABLE IF EXISTS users_new`,
		`CREATE TABLE users_new (
			id            TEXT PRIMARY KEY,
			username      TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO users_new (id, username, password_hash, created_at)
			SELECT id, ` + loginCol + `, password_hash, created_at FROM users`,
		`DROP TABLE users`,
		`ALTER TABLE users_new RENAME TO users`,
		`PRAGMA foreign_keys = ON`,
	}
	for _, s := range steps {
		if _, err = db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func migrate(database *sql.DB) error {
	if err := migrateUsersTable(database); err != nil {
		return err
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id            TEXT PRIMARY KEY,
			username      TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS auth_sessions (
			id         TEXT PRIMARY KEY,
			user_id    TEXT NOT NULL,
			token      TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS rcon_sessions (
			id         TEXT PRIMARY KEY,
			user_id    TEXT NOT NULL,
			name       TEXT NOT NULL,
			icon       TEXT NOT NULL DEFAULT '',
			host       TEXT NOT NULL,
			port       INTEGER NOT NULL,
			password   TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS command_history (
			id               TEXT PRIMARY KEY,
			rcon_session_id  TEXT NOT NULL,
			user_id          TEXT NOT NULL,
			command          TEXT NOT NULL,
			output           TEXT NOT NULL DEFAULT '',
			executed_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (rcon_session_id) REFERENCES rcon_sessions(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_auth_sessions_token     ON auth_sessions(token)`,
		`CREATE INDEX IF NOT EXISTS idx_command_history_session ON command_history(rcon_session_id, executed_at DESC)`,
	}
	for _, s := range stmts {
		if _, err := database.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
