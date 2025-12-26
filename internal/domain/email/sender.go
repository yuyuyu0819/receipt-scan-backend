package email

import "context"

// Sender sends confirmation emails to users.
type Sender interface {
	SendConfirmation(ctx context.Context, address string, userID int64) error
}
