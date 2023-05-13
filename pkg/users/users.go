package users

import (
	"context"
	"github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/authenticator"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/store"
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

type UserInformation struct {
	SpotifyToken       *oauth2.Token
	UserType           UserType
	UserName           string
	UserStartListening bool
}

type User struct {
	Identifier string
	userInfo   store.Store[UserInformation]
	player
}

func New(ctx context.Context, identifier string, username string, usertype UserType, store store.Store[UserInformation], token *oauth2.Token, auth *authenticator.Authenticator) (*User, error) {
	info := &UserInformation{
		UserType:           usertype,
		UserName:           username,
		SpotifyToken:       token,
		UserStartListening: false,
	}

	if err := store.Save(info, identifier); err != nil {
		return nil, err
	}

	return &User{
		Identifier: identifier,
		userInfo:   store,
		player:     NewPlayer(ctx, auth, token),
	}, nil
}

func NewEmpty() *User {
	return &User{
		Identifier: "",
		player:     player{},
	}
}

func (u *User) GetInfo() (*UserInformation, error) {
	if u.Identifier == "" {
		return &UserInformation{
			UserType:           UserTypeEmpty,
			UserName:           "",
			UserStartListening: false,
		}, nil
	} else {
		return u.userInfo.Get(u.Identifier)
	}
}

func (u *User) SetInfo(info *UserInformation) error {
	if u.Identifier == "" {
		return nil
	} else {
		return u.userInfo.Save(info, u.Identifier)
	}
}

func Load(ctx context.Context, identifier string, store store.Store[UserInformation], authenticator *authenticator.Authenticator) *User {
	info, _ := store.Get(identifier)
	return &User{
		Identifier: identifier,
		userInfo:   store,
		player:     NewPlayer(ctx, authenticator, info.SpotifyToken),
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
