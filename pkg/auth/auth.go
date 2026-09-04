package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string) string {
	pwd_hash := sha256.Sum256([]byte(password))
	pwd := hex.EncodeToString(pwd_hash[:])
	return pwd
}

func CheckPassword(password string, hash string) bool {
	return HashPassword(password) == hash
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	back := hex.EncodeToString(b)
	return back, nil
}
