package authenticator

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	*spotifyauth.Authenticator
}

func NewAuthenticator(redirectURL string, clientID string, secretKey string) *Authenticator {
	var scopes = []string{
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopeUserReadEmail,
		spotifyauth.ScopeUserModifyPlaybackState,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopePlaylistModifyPrivate,
		spotifyauth.ScopeImageUpload,
	}
	a := spotifyauth.New(spotifyauth.WithClientID(clientID), spotifyauth.WithClientSecret(secretKey), spotifyauth.WithRedirectURL(redirectURL), spotifyauth.WithScopes(scopes...))
	return &Authenticator{
		a,
	}
}

func (a *Authenticator) Authenticate(state string, r *http.Request) (*oauth2.Token, string, string, error) {
	token, err := a.Token(r.Context(), state, r)
	if err != nil {
		return nil, "", "", err
	}
	client := spotify.New(a.Client(r.Context(), token))
	user, err := client.CurrentUser(r.Context())
	if err != nil {
		return nil, "", "", err
	}
	hash := sha1.Sum([]byte(user.Email))
	return token, hex.EncodeToString(hash[:]), user.DisplayName, nil
}

func (a *Authenticator) CallbackURL(state string) string {
	return a.AuthURL(state)
}
