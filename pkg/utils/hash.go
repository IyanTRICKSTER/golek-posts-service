package utils

import "golang.org/x/crypto/bcrypt"

func Hash(data string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(data), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}

	return string(hashed)
}

func HashCompare(hashed string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(hash))
	if err != nil {
		return false
	}
	return true
}
