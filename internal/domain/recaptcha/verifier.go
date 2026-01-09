package recaptcha

import "context"

// Verifier validates reCAPTCHA tokens.
type Verifier interface {
	Verify(ctx context.Context, token string) error
}
