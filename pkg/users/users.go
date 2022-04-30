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
	SpotifyToken *oauth2.Token
	UserType     UserType
	UserName     string
}

type User struct {
	Identifier string
	userInfo   store.Store[UserInformation]
	player
}

func New(identifier string, username string, usertype UserType, store store.Store[UserInformation], token *oauth2.Token, auth *authenticator.Authenticator) *User {
	info := &UserInformation{
		UserType:     usertype,
		UserName:     username,
		SpotifyToken: token,
	}

	store.Save(info, identifier)

	return &User{
		Identifier: identifier,
		userInfo:   store,
		player:     NewPlayer(auth, token),
	}
}

func NewEmpty() *User {
	return &User{
		Identifier: "",
		player:     player{},
	}
}

func (u *User) GetInfo() *UserInformation {
	if u.Identifier == "" {
		return &UserInformation{
			UserType: UserTypeEmpty,
			UserName: "",
		}
	} else {
		info, _ := u.userInfo.Get(u.Identifier)
		return info
	}
}

func (u *User) SetInfo(info *UserInformation) {
	if u.Identifier == "" {
		return
	} else {
		u.userInfo.Save(info, u.Identifier)
		return
	}
}

func Load(identifier string, store store.Store[UserInformation], authenticator *authenticator.Authenticator) *User {
	info, _ := store.Get(identifier)
	return &User{
		Identifier: identifier,
		userInfo:   store,
		player:     NewPlayer(authenticator, info.SpotifyToken),
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
