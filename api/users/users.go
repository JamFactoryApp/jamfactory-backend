package users

import (
	"context"
	"github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
)

type contextKey string

const key contextKey = "UserIdentifier"

func NewContext(ctx context.Context, user *types.User) context.Context {
	return context.WithValue(ctx, key, user)
}

func FromContext(ctx context.Context) (*types.User, error) {
	val := ctx.Value(key)
	if val == nil {
		return nil, errors.ErrSessionMissing
	}
	session, ok := val.(*types.User)
	if !ok {
		return nil, errors.ErrSessionMalformed
	}
	return session, nil
}
