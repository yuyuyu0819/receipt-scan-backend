package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"receiptScan-backend/internal/usecase/token"
)

type TokenHandler struct {
	tokenUsecase token.UseCase
}

func NewTokenHandler(t token.UseCase) *TokenHandler {
	return &TokenHandler{tokenUsecase: t}
}

type tokenRefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type tokenRefreshResponse struct {
	Token string `json:"token"`
}

// ServeRefresh は POST /api/token/refresh を処理します。
// refreshToken を検証し、新しい accessToken を返します。
func (h *TokenHandler) ServeRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req tokenRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "refreshToken is required", http.StatusBadRequest)
		return
	}

	out, err := h.tokenUsecase.Refresh(r.Context(), token.RefreshInput{Token: req.RefreshToken})
	if err != nil {
		if errors.Is(err, token.ErrInvalidToken) || errors.Is(err, token.ErrTokenExpired) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tokenRefreshResponse{Token: accessToken})
}

type tokenRevokeRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// ServeRevoke は POST /api/token/revoke を処理します。
// refreshToken を DB から削除してログアウトします。
func (h *TokenHandler) ServeRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req tokenRevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		http.Error(w, "refreshToken is required", http.StatusBadRequest)
		return
	}

	if err := h.tokenUsecase.Revoke(r.Context(), token.RevokeInput{Token: req.RefreshToken}); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
