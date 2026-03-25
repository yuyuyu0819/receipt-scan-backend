package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"receiptScan-backend/internal/usecase/login"
	"receiptScan-backend/internal/usecase/token"
)

type LoginHandler struct {
	usecase      login.UseCase
	tokenUsecase token.UseCase
}

func NewLoginHandler(u login.UseCase, t token.UseCase) *LoginHandler {
	return &LoginHandler{usecase: u, tokenUsecase: t}
}

type loginRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type loginUser struct {
	ID       int64  `json:"id"`
	UserName string `json:"userName"`
}

type loginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
	User         loginUser `json:"user"`
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	out, err := h.usecase.Execute(r.Context(), login.Input{UserName: req.UserName, Password: req.Password})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, login.ErrInvalidUserName),
			errors.Is(err, login.ErrInvalidPassword):
			status = http.StatusBadRequest
		case errors.Is(err, login.ErrAuthenticationFailed):
			status = http.StatusUnauthorized
		}
		http.Error(w, err.Error(), status)
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		http.Error(w, "JWT_SECRET is not set", http.StatusInternalServerError)
		return
	}
	accessToken, err := buildJWT(secret, out.UserID, time.Now(), 24*time.Hour)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	tokenOut, err := h.tokenUsecase.Create(r.Context(), token.CreateInput{UserID: out.UserID})
	if err != nil {
		http.Error(w, "failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{
		Token:        accessToken,
		RefreshToken: tokenOut.Token,
		User:         loginUser{ID: out.UserID, UserName: req.UserName},
	})
}
