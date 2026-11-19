// Package storage persists high scores and settings in a local SQLite file.
package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	_ "modernc.org/sqlite" // pure Go driver, no CGO

	"github.com/rjeff-sudo/termsnake-go/migrations"
)

// Score is one saved run.
type Score struct {
	ID         int64
	PlayerName string
	Score      int
	Difficulty string
	SnakeLen   int
	DurationMS int64
	CreatedAt  string
}

// Store owns the SQLite connection.
type Store struct {
	db *sql.DB
}

// DefaultPath returns the per-user database location. On Linux that follows
// the XDG data directory, e.g. ~/.local/share/termsnake/termsnake.db.
func DefaultPath() (string, error) {
	if runtime.GOOS == "linux" {
		base := os.Getenv("XDG_DATA_HOME")
		if base == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			base = filepath.Join(home, ".local", "share")
		}
		return filepath.Join(base, "termsnake", "termsnake.db"), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "termsnake", "termsnake.db"), nil
}

// Open creates the parent directory, opens the database and applies the
// embedded schema.
func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("storage: empty database path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("storage: create directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage: open: %w", err)
	}
	// One connection keeps the pure-Go driver out of locking trouble; the game
	// is the only writer anyway.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec("PRAGMA busy_timeout = 5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: busy_timeout: %w", err)
	}
	// WAL is a nice-to-have; some filesystems refuse it, which is fine.
	_, _ = db.Exec("PRAGMA journal_mode = WAL")

	if _, err := db.Exec(migrations.Init); err != nil {
		db.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

// SaveScore inserts a finished run.
func (s *Store) SaveScore(score Score) error {
	name := score.PlayerName
	if name == "" {
		name = "PLAYER"
	}
	_, err := s.db.Exec(
		`INSERT INTO scores (player_name, score, difficulty, snake_len, duration_ms)
		 VALUES (?, ?, ?, ?, ?)`,
		name, score.Score, score.Difficulty, score.SnakeLen, score.DurationMS,
	)
	if err != nil {
		return fmt.Errorf("storage: save score: %w", err)
	}
	return nil
}

// TopScores returns the best runs, highest score first and earliest run first
// on ties. Pass an empty difficulty for all difficulties.
func (s *Store) TopScores(difficulty string, limit int) ([]Score, error) {
	if limit <= 0 {
		limit = 10
	}
	query := `SELECT id, player_name, score, difficulty, snake_len, duration_ms, created_at
	          FROM scores`
	args := make([]any, 0, 2)
	if difficulty != "" {
		query += ` WHERE difficulty = ?`
		args = append(args, difficulty)
	}
	query += ` ORDER BY score DESC, created_at ASC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("storage: top scores: %w", err)
	}
	defer rows.Close()

	var scores []Score
	for rows.Next() {
		var sc Score
		if err := rows.Scan(&sc.ID, &sc.PlayerName, &sc.Score, &sc.Difficulty,
			&sc.SnakeLen, &sc.DurationMS, &sc.CreatedAt); err != nil {
			return nil, fmt.Errorf("storage: scan score: %w", err)
		}
		scores = append(scores, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage: read scores: %w", err)
	}
	return scores, nil
}

// GetSetting returns the stored value, or "" when the key is absent.
func (s *Store) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("storage: get setting: %w", err)
	}
	return value, nil
}

// SetSetting stores or replaces a setting.
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("storage: set setting: %w", err)
	}
	return nil
}
