package user

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashPassword hashes a password using user-specific data.
func HashPassword(userID string, password string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", userID, password)))
	return hex.EncodeToString(sum[:])
}
