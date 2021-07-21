package types

import "golang.org/x/oauth2"

type UserType string

const (
	UserTypeSession UserType = "Session"
	UserTypeSpotify UserType = "Spotify"
)

type User struct {
	Identifier string
	UserType   UserType
	UserName   string
	Token      *oauth2.Token
}
