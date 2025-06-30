package adapters

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher defines the interface for password hashing operations.
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

// BcryptHasher implements PasswordHasher using bcrypt.
type BcryptHasher struct{}

func NewPasswordHasher() PasswordHasher {
	return &BcryptHasher{}
}

func (b *BcryptHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (b *BcryptHasher) CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
