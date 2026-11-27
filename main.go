package main

import (
	"auth-api/config"
	"auth-api/database"
	"auth-api/handlers"
	"auth-api/models"
	"auth-api/pkg/auth"
	"auth-api/pkg/middleware"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	users := models.NewUserStore(db)
	tokenStore := models.NewTokenStore(db)
	bookStore := models.NewBookStore(db)
	tokens := auth.NewTokenManager(cfg.JWTSecret, 15*time.Minute, 24*time.Hour)
	authHandler := handlers.NewAuthHandler(users, tokens, tokenStore)
	bookHandler := handlers.NewBookHandler(bookStore)
	requireAuth := middleware.Auth(tokens, tokenStore)
	mux := http.NewServeMux()
	registerAuthRoutes(mux, authHandler, requireAuth)
	registerBookRoutes(mux, bookHandler, requireAuth)

	log.Println("listening on", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(cfg.Port), mux))
}
func registerAuthRoutes(mux *http.ServeMux, h *handlers.AuthHandler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/refresh", h.Refresh)
	mux.HandleFunc("POST /auth/logout", requireAuth(h.Logout))
	mux.HandleFunc("GET /auth/me", requireAuth(h.Me))
}

func registerBookRoutes(mux *http.ServeMux, h *handlers.BookHandler, requireAuth func(http.HandlerFunc) http.HandlerFunc) {
	mux.HandleFunc("GET /books", requireAuth(h.List))
	mux.HandleFunc("GET /books/{id}", requireAuth(h.Get))
	mux.HandleFunc("POST /books", requireAuth(h.Create))
	mux.HandleFunc("DELETE /books/{id}", requireAuth(h.Delete))
}
