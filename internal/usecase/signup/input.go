package signup

// Input represents a signup request.
type Input struct {
	UserID         int64
	Password       string
	Email          string
	RecaptchaToken string
}
