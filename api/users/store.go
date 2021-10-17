package users

import (
	"errors"
)

var (
	ErrUserNotFound     = errors.New("UserStore: user not found")
	ErrInterfaceConvert = errors.New("RedisUserStore: Failed to convert user from interface{} to []bytes")
)

// Store stores users
type Store interface {
	// Get returns the user with the provided identifier
	Get(identifier string) (*User, error)
	// Save stores the user in the store
	Save(user *User) error
	// Delete deletes the user with the provided identifier
	Delete(identifier string) error
}
