package utils

import (
	"errors"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func ParseTokenFromSession(session *sessions.Session) (*oauth2.Token, error) {
	if token, ok := session.Values[SessionTokenKey].(*oauth2.Token); ok {
		return token, nil
	}
	return nil, errors.New("TokenParser: Failed to parse token from session")
}
