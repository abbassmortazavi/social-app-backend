package handlers

import (
	"auth-api/models"
	"auth-api/pkg/middleware"
	"auth-api/pkg/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type BookHandler struct {
	books *models.BookStore
}

func NewBookHandler(books *models.BookStore) *BookHandler {
	return &BookHandler{
		books: books,
	}
}

type CreateBookRequest struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

func (b *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Description == "" {
		utils.WriteError(w, http.StatusBadRequest, "description is required")
		return
	}
	if req.Title == "" {
		utils.WriteError(w, http.StatusBadRequest, "title is required")
		return
	}
	if req.Author == "" {
		utils.WriteError(w, http.StatusBadRequest, "author is required")
		return
	}
	defer r.Body.Close()
	existing, err := b.books.FindByTitle(req.Title)
	if err != nil && !errors.Is(err, models.ErrBookNotFound) {
		utils.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if existing != nil {
		utils.WriteError(w, http.StatusBadRequest, "Book already exists")
		return
	}

	user, ok := middleware.UserFrom(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	book := &models.Book{
		Title:       req.Title,
		Author:      req.Author,
		Description: req.Description,
		User: models.User{
			ID: user.ID,
		},
	}
	err = b.books.Create(book)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJson(w, http.StatusCreated, "Book created")
}

type ListBooksResponse struct {
	Books []*models.Book `json:"books"`
}
type ListBookResponse struct {
	Book *models.Book `json:"book"`
}

func (b *BookHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFrom(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	lists, err := b.books.Lists(user.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := &ListBooksResponse{
		Books: lists,
	}
	utils.WriteJson(w, http.StatusOK, resp)
}
func (b *BookHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetPathParamInt(r, "id")
	book, err := b.books.FindByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := &ListBookResponse{
		Book: book,
	}
	utils.WriteJson(w, http.StatusOK, resp)
}
func (b *BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := utils.GetPathParamInt(r, "id")
	log.Println("id:", id)
	book, err := b.books.FindByID(id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = b.books.DeleteByID(book.ID)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	utils.WriteJson(w, http.StatusOK, "Book deleted")
}
