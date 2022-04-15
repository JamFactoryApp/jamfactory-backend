package users

import (
	"errors"
)

var (
	ErrUserNotFound     = errors.New("Store: user not found")
	ErrInterfaceConvert = errors.New("Store: Failed to convert user from interface{} to []bytes")
)
