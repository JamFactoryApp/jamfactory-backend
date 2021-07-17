package server

import (
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"golang.org/x/oauth2"
	"net/http"
)

// JamFactory provides methods to control JamSessions
type JamFactory interface {
	// Authenticate takes a request and state and returns an OAuth2 token
	Authenticate(state string, r *http.Request) (*oauth2.Token, error)
	// CallbackURL returns a URL a user should visit for authentication
	CallbackURL(state string) string
	// DeleteJamSession deletes the JamSession with the given jamLabel if it exists
	DeleteJamSession(jamLabel string) error
	// GetJamSession returns the JamSession for a given jamLabel
	GetJamSession(jamLabel string) (jamsession.JamSession, error)
	// NewJamSession creates a new JamSession using the user account provided by the OAuth2 token
	NewJamSession(*oauth2.Token) (jamsession.JamSession, error)
	// Search yields search results from the music streaming provider
	Search(jamSession jamsession.JamSession, t string, text string) (interface{}, error)
	// ClientAddress returns the address this JamFactory's client listens on
	ClientAddress() string
}
