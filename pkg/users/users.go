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
	Identifier string
	UserType   UserType
	UserName   string
	Player
}

func New(identifier string, username string, usertype UserType, token *oauth2.Token, auth *Authenticator) *User {
	return &User{
		Identifier: identifier,
		UserType:   usertype,
		UserName:   username,
		Player: Player{
			authenticator: auth,
			SpotifyToken:  token,
		},
	}
}

func NewEmpty() *User {
	return &User{
		Identifier: "",
		UserType:   UserTypeEmpty,
		UserName:   "",
		Player:     Player{},
	}
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
