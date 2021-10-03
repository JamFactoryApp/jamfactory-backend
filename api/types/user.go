package types

import "golang.org/x/oauth2"

type UserType string

const (
	UserTypeEmpty   UserType = "Empty"
	UserTypeSession UserType = "Session"
	UserTypeSpotify UserType = "Spotify"
)

type User struct {
	Identifier string
	UserType   UserType
	UserName     string
	SpotifyToken *oauth2.Token
}
