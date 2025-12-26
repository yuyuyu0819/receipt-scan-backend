package http

import (
	"net/http"
)

// ActivationHandler returns a simple activation completion page.
type ActivationHandler struct{}

func NewActivationHandler() *ActivationHandler {
	return &ActivationHandler{}
}

func (h *ActivationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<html><body><h1>Activation complete</h1></body></html>"))
}
