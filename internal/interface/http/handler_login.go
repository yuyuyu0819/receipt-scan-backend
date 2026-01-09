package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"receiptScan-backend/internal/usecase/login"
)

type LoginHandler struct {
	usecase login.UseCase
}

func NewLoginHandler(u login.UseCase) *LoginHandler {
	return &LoginHandler{usecase: u}
}

type loginRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type loginResponse struct {
	Message string `json:"message"`
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

	_, err := h.usecase.Execute(r.Context(), login.Input{UserName: req.UserName, Password: req.Password})
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(loginResponse{Message: "ok"})
}
