package jamsession

import (
	"errors"
)

var (
	ErrJamNotFound      = errors.New("JamStore: jamSession not found")
	ErrInterfaceConvert = errors.New("Store: Failed to convert user from interface{} to []bytes")
)
