package user

import (
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"golang.org/x/oauth2"
)

type User struct {
	Identifier string
	UserType   types.UserType
	UserName   string
	Token      *oauth2.Token
}

func NewUser(identifier string, username string, usertype types.UserType, token *oauth2.Token) *User {
	return &User{
		Identifier: identifier,
		UserType:   usertype,
		UserName:   username,
		Token:      token,
	}
}
