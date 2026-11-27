package handlers

import (
	"auth-api/models"
	"auth-api/pkg/auth"
	"auth-api/pkg/middleware"
	"auth-api/pkg/utils"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	_ "golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	users      *models.UserStore
	tokens     *auth.TokenManager
	tokenStore *models.TokenStore
}
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type UserResponse struct {
	Token TokenResponse `json:"token"`
	User  interface{}   `json:"user"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func NewAuthHandler(users *models.UserStore, tokens *auth.TokenManager, tokenStore *models.TokenStore) *AuthHandler {
	return &AuthHandler{
		users:      users,
		tokens:     tokens,
		tokenStore: tokenStore,
	}
}

type CredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
}

func (h *AuthHandler) issueTokens(user *models.User) (*TokenResponse, error) {
	access, eccessEx, err := h.tokens.GenerateAccessToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}
	refresh, refreshEx, err := h.tokens.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	err = h.tokenStore.ReplaceForUser(user.ID,
		models.NewToken{Type: auth.Access, Hash: auth.HashToken(access), ExpiresAt: eccessEx},
		models.NewToken{Type: auth.Refresh, Hash: auth.HashToken(refresh), ExpiresAt: refreshEx},
	)
	if err != nil {
		return nil, err
	}
	return &TokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int64(h.tokens.AccessTTL),
	}, nil
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var c CredentialsRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1048576)).Decode(&c); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if c.Email == "" || c.Password == "" {
		utils.WriteError(w, http.StatusBadRequest, "Missing email or password")
		return
	}
	c.Email = strings.TrimSpace(strings.ToLower(c.Email))
	user, err := h.users.FindByEmail(c.Email)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(c.Password)); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	token, err := h.issueTokens(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tk := &TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    int64(h.tokens.AccessTTL),
	}
	userResponse := UserResponse{
		Token: *tk,
		User:  user,
	}
	utils.WriteJson(w, http.StatusOK, userResponse)
}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var c CredentialsRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1048576)).Decode(&c); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	c.Email = strings.TrimSpace(strings.ToLower(c.Email))
	if c.Email == "" || len(c.Password) < 6 {
		utils.WriteError(w, http.StatusBadRequest, "email required and password must be at least 6 chars")
		return
	}
	user := &models.User{}
	hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user.Password = string(hash)
	user.Email = c.Email
	user.Name = c.Name
	user.Bio = c.Bio
	err = h.users.Create(user)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	getUser, err := h.users.FindByEmail(c.Email)
	token, err := h.issueTokens(getUser)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tk := &TokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    int64(h.tokens.AccessTTL),
	}

	userResponse := &UserResponse{
		Token: *tk,
		User:  getUser,
	}

	log.Println("Registered user:", user)
	utils.WriteJson(w, http.StatusOK, userResponse)
}
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshTokenRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil || req.RefreshToken == "" {
		utils.WriteError(w, http.StatusBadRequest, "invalid body")
		return
	}

	userID, err := h.tokenStore.FindUserIDByRefresh(auth.HashToken(req.RefreshToken))
	if err != nil {
		if errors.Is(err, models.ErrTokenNotFound) {
			utils.WriteError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		}
		log.Printf("refresh: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "server error")
		return
	}

	user, err := h.users.FindByID(userID)
	if err != nil {
		utils.WriteError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	pair, err := h.issueTokens(user)
	if err != nil {
		log.Printf("refresh: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "server error")
		return
	}
	utils.WriteJson(w, http.StatusOK, pair)
}

type UserMeData struct {
	User models.User `json:"user"`
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFrom(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	userMe := UserMeData{
		User: models.User{
			Email: user.Email,
			Name:  user.Name,
			ID:    user.ID,
		},
	}
	utils.WriteJson(w, http.StatusOK, userMe)
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFrom(r.Context())
	if !ok {
		utils.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.tokenStore.DeleteAllForUser(user.ID); err != nil {
		log.Printf("logout: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, "server error")
		return
	}
	utils.WriteJson(w, http.StatusOK, map[string]string{"message": "logged out"})
}
