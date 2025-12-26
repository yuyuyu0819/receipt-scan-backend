package recaptcha

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	domain "receiptScan-backend/internal/domain/recaptcha"
)

// GoogleVerifier verifies reCAPTCHA tokens against Google.
type GoogleVerifier struct {
	secret     string
	httpClient *http.Client
}

// NewGoogleVerifierFromEnv builds a verifier using RECAPTCHA_SECRET.
func NewGoogleVerifierFromEnv() (*GoogleVerifier, error) {
	secret := strings.TrimSpace(os.Getenv("RECAPTCHA_SECRET"))
	if secret == "" {
		return nil, errors.New("RECAPTCHA_SECRET must be set")
	}

	return &GoogleVerifier{
		secret: secret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

type verifyResponse struct {
	Success    bool     `json:"success"`
	ErrorCodes []string `json:"error-codes"`
}

// Verify validates the token using Google's siteverify endpoint.
func (v *GoogleVerifier) Verify(ctx context.Context, token string) error {
	form := url.Values{}
	form.Set("secret", v.secret)
	form.Set("response", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.google.com/recaptcha/api/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("verify request: %w", err)
	}
	defer resp.Body.Close()

	var payload verifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if !payload.Success {
		if len(payload.ErrorCodes) > 0 {
			return fmt.Errorf("verification failed: %v", payload.ErrorCodes)
		}
		return errors.New("verification failed")
	}

	return nil
}

var _ domain.Verifier = (*GoogleVerifier)(nil)
