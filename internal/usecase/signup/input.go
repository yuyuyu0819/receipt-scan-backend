package signup

// Input represents a signup request.
type Input struct {
	UserName       string
	Password       string
	Email          string
	RecaptchaToken string
}
