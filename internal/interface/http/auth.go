package http

import "context"

type contextKey string

const userIDKey contextKey = "userID"

func withUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func userIDFromContext(ctx context.Context) (int64, bool) {
	value := ctx.Value(userIDKey)
	userID, ok := value.(int64)
	return userID, ok
}
