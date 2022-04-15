package jamsession

import (
	"context"
	"github.com/pkg/errors"
)

var (
	ErrJamSessionMissing   = errors.New("no JamSession provided")
	ErrJamSessionMalformed = errors.New("malformed JamSession")
)

type contextKey string

const key contextKey = "JamSession"

// NewContext returns a new context containing a JamSession
func NewContext(ctx context.Context, jamSession *JamSession) context.Context {
	return context.WithValue(ctx, key, jamSession)
}

// FromContext returns a JamSession existing in a context
func FromContext(ctx context.Context) (*JamSession, error) {
	val := ctx.Value(key)
	if val == nil {
		return nil, ErrJamSessionMissing
	}
	jamSession, ok := val.(*JamSession)
	if !ok {
		return nil, ErrJamSessionMalformed
	}
	return jamSession, nil
}
