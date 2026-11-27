package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrEmailExists = errors.New("email already exists")
	ErrNotFound    = errors.New("user not found")
)

type User struct {
	ID          int            `json:"id"`
	Email       string         `json:"email"`
	Password    string         `json:"-"`
	Name        string         `json:"name"`
	Image       sql.NullString `json:"image,omitempty"`
	Bio         string         `json:"bio,omitempty"`
	Address     string         `json:"address,omitempty"`
	PhoneNumber string         `json:"phone_number,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db}
}
func (s *UserStore) Create(user *User) error {
	query := `INSERT INTO users (email, password, name, image, bio, address, phone_number) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := s.db.Exec(query, user.Email, user.Password, user.Name, user.Image, user.Bio, user.Address, user.PhoneNumber)
	if err != nil {
		if pqErr, ok := errors.AsType[*pq.Error](err); ok && pqErr.Code == "23505" {
			return ErrEmailExists
		}
		return err
	}
	return nil
}
func (s *UserStore) FindByEmail(email string) (*User, error) {
	query := `SELECT id, email, password, name, image, bio, address, phone_number, created_at FROM users WHERE email = $1`
	row := s.db.QueryRow(query, email)
	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Image,
		&user.Bio,
		&user.Address,
		&user.PhoneNumber,
		&user.CreatedAt)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
func (s *UserStore) FindByID(id int) (*User, error) {
	query := `SELECT id, email, password, name, image,bio, address, phone_number, created_at FROM users WHERE id = $1`
	row := s.db.QueryRow(query, id)
	user := &User{}
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Image,
		&user.Bio,
		&user.Address,
		&user.PhoneNumber,
		&user.CreatedAt)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
