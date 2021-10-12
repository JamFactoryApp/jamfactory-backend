package users

import (
	"context"
	"github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"golang.org/x/oauth2"
)

type UserType string
type contextKey string

const key contextKey = "User"

const (
	UserTypeEmpty   UserType = "Empty"
	UserTypeSession UserType = "Session"
	UserTypeSpotify UserType = "Spotify"
)

type User struct {
	Identifier   string
	UserType     UserType
	UserName     string
	SpotifyToken *oauth2.Token
}

func NewContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, key, user)
}

func FromContext(ctx context.Context) (*User, error) {
	val := ctx.Value(key)
	if val == nil {
		return nil, errors.ErrSessionMissing
	}
	session, ok := val.(*User)
	if !ok {
		return nil, errors.ErrSessionMalformed
	}
	return session, nil
}
