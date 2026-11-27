package models

import (
	"database/sql"
	"errors"
	"time"
)

var ErrTokenNotFound = errors.New("token not found")

type TokenStore struct {
	db *sql.DB
}

func NewTokenStore(db *sql.DB) *TokenStore {
	return &TokenStore{
		db: db,
	}
}

type NewToken struct {
	Type      string    `json:"type"`
	Hash      string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (s *TokenStore) ReplaceForUser(userID int, tokens ...NewToken) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec("DELETE FROM tokens WHERE user_id = $1", userID); err != nil {
		return err
	}
	for _, token := range tokens {
		_, err := tx.Exec("INSERT INTO tokens (user_id, token_type, hash_token, expired_at) VALUES ($1, $2, $3, $4)", userID, token.Type, token.Hash, token.ExpiresAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *TokenStore) IsValid(tokenType, hash string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM tokens
			WHERE hash_token = $1 AND token_type = $2
			  AND is_revoked = FALSE AND expired_at > $3
		)`,
		hash, tokenType, time.Now().UTC(),
	).Scan(&exists)
	return exists, err
}
func (s *TokenStore) FindUserIDByRefresh(hash string) (int, error) {
	var userID int
	err := s.db.QueryRow(
		`SELECT user_id FROM tokens
		 WHERE hash_token = $1 AND token_type = 'refresh'
		   AND is_revoked = FALSE AND expired_at > $2`,
		hash, time.Now().UTC(),
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrTokenNotFound
	}
	return userID, err
}

func (s *TokenStore) DeleteAllForUser(userID int) error {
	_, err := s.db.Exec(`DELETE FROM tokens WHERE user_id = $1`, userID)
	return err
}
