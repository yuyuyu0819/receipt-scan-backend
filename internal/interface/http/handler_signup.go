package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"receiptScan-backend/internal/usecase/signup"
)

type SignupHandler struct {
	usecase signup.UseCase
}

func NewSignupHandler(u signup.UseCase) *SignupHandler {
	return &SignupHandler{usecase: u}
}

type signupRequest struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type signupResponse struct {
	Message string `json:"message"`
}

func (h *SignupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err := h.usecase.Execute(r.Context(), signup.Input{
		UserName: req.UserName,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, signup.ErrInvalidUserName),
			errors.Is(err, signup.ErrInvalidPassword),
			errors.Is(err, signup.ErrInvalidEmail):
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(signupResponse{Message: "created"})
}
