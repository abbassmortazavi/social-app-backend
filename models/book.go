package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrBookNotFound = errors.New("book not found")
)

type Book struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
	User        User      `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
}

type BookStore struct {
	db *sql.DB
}

func NewBookStore(db *sql.DB) *BookStore {
	return &BookStore{db}
}
func (s *BookStore) Create(book *Book) error {
	query := `INSERT INTO books (title, description, author, user_id) VALUES ($1, $2, $3, $4)`
	_, err := s.db.Exec(query, book.Title, book.Description, book.Author, book.User.ID)
	if err != nil {
		return err
	}
	return nil
}
func (b *BookStore) FindByUserID(userID int64) (*Book, error) {
	query := `SELECT id, title, description, author, user_id, created_at FROM books WHERE user_id = $1`
	row := b.db.QueryRow(query, userID)
	book := &Book{}
	err := row.Scan(&book.ID, &book.Title, &book.Description, &book.Author, &book.User.ID, &book.CreatedAt)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return book, nil
}
func (b *BookStore) FindByID(id int) (*Book, error) {
	query := `SELECT b.id, b.title, b.description, b.author, b.created_at ,
		 u.id, u.name, u.email, u.created_at
FROM books b 
	JOIN users u ON b.user_id = u.id
                                                                                                           
    WHERE b.id = $1`
	row := b.db.QueryRow(query, id)
	book := &Book{}
	err := row.Scan(&book.ID, &book.Title, &book.Description, &book.Author, &book.CreatedAt, &book.User.ID, &book.User.Name, &book.User.Email, &book.User.CreatedAt)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return book, nil
}
func (b *BookStore) FindByTitle(title string) (*Book, error) {
	query := `SELECT id, title, description, author, user_id, created_at FROM books WHERE title = $1`
	row := b.db.QueryRow(query, title)
	book := &Book{}
	err := row.Scan(&book.ID, &book.Title, &book.Description, &book.Author, &book.User.ID, &book.CreatedAt)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrBookNotFound
		}
		return nil, err
	}
	return book, nil
}
func (b *BookStore) DeleteByID(id int) error {
	query := `DELETE FROM books WHERE id = $1`
	_, err := b.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
func (b *BookStore) Lists(userID int) ([]*Book, error) {
	query := `SELECT b.id, b.title, b.description, b.author, b.created_at,
       u.id, u.name, u.email, u.created_at
	FROM books b
    JOIN users u ON b.user_id = u.id   
	where b.user_id = $1
	order by b.created_at desc`
	rows, err := b.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var books []*Book
	for rows.Next() {
		book := &Book{}
		err := rows.Scan(&book.ID, &book.Title, &book.Description, &book.Author, &book.CreatedAt, &book.User.ID, &book.User.Name, &book.User.Email, &book.User.CreatedAt)
		if err != nil {
			return nil, err
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return books, nil

}
