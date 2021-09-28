package types

import "golang.org/x/oauth2"

type UserType string

type contextKey string

const key contextKey = "UserIdentifier"

const (
	UserTypeEmpty   UserType = "Empty"
	UserTypeSession UserType = "Session"
	UserTypeSpotify UserType = "Spotify"
)

type User struct {
	Identifier string
	UserType   UserType
	UserName   string
	Token      *oauth2.Token
}
