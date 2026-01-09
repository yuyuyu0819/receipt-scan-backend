package user

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// HashPassword hashes a password using user-specific data.
func HashPassword(userName string, password string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", userName, password)))
	return hex.EncodeToString(sum[:])
}
