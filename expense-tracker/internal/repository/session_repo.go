package repository

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"expense-tracker/internal/models"
)

var ErrSessionNotFound = errors.New("сессия не найдена")

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create генерирует случайный session_id и сохраняет сессию.
func (r *SessionRepository) Create(userID int64, ttl time.Duration) (*models.Session, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	sessionID := hex.EncodeToString(buf)
	expiresAt := time.Now().Add(ttl).UTC().Format(time.RFC3339)

	_, err := r.db.Exec(
		"INSERT INTO sessions (session_id, user_id, expires_at) VALUES (?, ?, ?)",
		sessionID, userID, expiresAt,
	)
	if err != nil {
		return nil, err
	}
	return &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

// GetValid возвращает сессию, если она существует и не истекла.
func (r *SessionRepository) GetValid(sessionID string) (*models.Session, error) {
	row := r.db.QueryRow(
		"SELECT session_id, user_id, expires_at FROM sessions WHERE session_id = ?",
		sessionID,
	)
	var s models.Session
	if err := row.Scan(&s.SessionID, &s.UserID, &s.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	exp, err := time.Parse(time.RFC3339, s.ExpiresAt)
	if err != nil || time.Now().After(exp) {
		_ = r.Delete(sessionID)
		return nil, ErrSessionNotFound
	}
	return &s, nil
}

func (r *SessionRepository) Delete(sessionID string) error {
	_, err := r.db.Exec("DELETE FROM sessions WHERE session_id = ?", sessionID)
	return err
}

// DeleteExpired — вызывает очистку старых сессий (можно дергать периодически).
func (r *SessionRepository) DeleteExpired() error {
	_, err := r.db.Exec(
		"DELETE FROM sessions WHERE expires_at < ?",
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}