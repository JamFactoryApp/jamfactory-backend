package users

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
	"net/http"
)

type Authenticator struct {
	spotify.Authenticator
}

func NewAuthenticator(redirectURL string, clientID string, secretKey string) *Authenticator {
	var scopes = []string{
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState,
		spotify.ScopePlaylistModifyPrivate,
		spotify.ScopeImageUpload,
	}
	a := spotify.NewAuthenticator(redirectURL, scopes...)
	a.SetAuthInfo(clientID, secretKey)
	return &Authenticator{
		a,
	}
}

func (a *Authenticator) Authenticate(state string, r *http.Request) (*oauth2.Token, string, string, error) {
	token, err := a.Token(state, r)
	if err != nil {
		return nil, "", "", err
	}
	client := a.NewClient(token)
	user, err := client.CurrentUser()
	if err != nil {
		return nil, "", "", err
	}
	hash := sha1.Sum([]byte(user.Email))
	return token, hex.EncodeToString(hash[:]), user.DisplayName, nil
}

func (a *Authenticator) CallbackURL(state string) string {
	return a.AuthURL(state)
}
