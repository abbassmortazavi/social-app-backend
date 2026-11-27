package database

import (
	"database/sql"
)

func Connect(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT  NULL,
			image TEXT  NULL,
			password TEXT NOT NULL,
			bio TEXT NULL,
			address TEXT NULL,
			phone_number TEXT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS tokens (
			id         SERIAL PRIMARY KEY,
			user_id    INTEGER     NOT NULL,
			token_type VARCHAR(20) NOT NULL,
			hash_token TEXT        NOT NULL,
			expired_at TIMESTAMP   NOT NULL,
			is_revoked BOOLEAN   DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS books (
			id         SERIAL PRIMARY KEY,
			user_id    INTEGER     NOT NULL,
			title VARCHAR(255) NOT NULL,
			author VARCHAR(255)        NOT NULL,
    		description text        NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			body TEXT  NULL,
			file TEXT  NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
    		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			text TEXT  NULL,
			file TEXT  NULL,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
    		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    		FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS post_likes (
			id SERIAL PRIMARY KEY,
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
    		FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    		FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			sender_id INTEGER NOT NULL,
			receiver_id INTEGER NOT NULL,
			title TEXT  NULL,
			data TEXT  NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
    		FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE,
    		FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE CASCADE
		)`,

		`CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_expired_at ON tokens (expired_at)`,
		`CREATE INDEX IF NOT EXISTS idx_tokens_hash_token ON tokens (hash_token)`,
		`CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments (user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_post_likes_post_id ON post_likes (post_id)`,
		`CREATE INDEX IF NOT EXISTS idx_post_likes_user_id ON post_likes (user_id)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}
