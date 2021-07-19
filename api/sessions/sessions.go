package sessions

import (
	"context"
	"github.com/gorilla/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"golang.org/x/oauth2"
)

type contextKey string

const key contextKey = "Session"

const (
	jamLabelKey    = "Label"
	tokenKey       = "CurrentToken"
	sessionTypeKey = "User"
	identifierKey  = "Identifier"
	originKey      = "Origin"
)

func NewContext(ctx context.Context, session *sessions.Session) context.Context {
	return context.WithValue(ctx, key, session)
}

func FromContext(ctx context.Context) (*sessions.Session, error) {
	val := ctx.Value(key)
	if val == nil {
		return nil, errors.ErrSessionMissing
	}
	session, ok := val.(*sessions.Session)
	if !ok {
		return nil, errors.ErrSessionMalformed
	}
	return session, nil
}

func JamLabel(session *sessions.Session) (string, error) {
	jamLabelVal := session.Values[jamLabelKey]
	if jamLabelVal == nil {
		return "", errors.ErrJamLabelMissing
	}

	jamLabel, ok := jamLabelVal.(string)
	if !ok {
		return "", errors.ErrJamLabelMalformed
	}

	return jamLabel, nil
}

func Token(session *sessions.Session) (*oauth2.Token, error) {
	tokenVal := session.Values[tokenKey]
	if tokenVal == nil {
		return nil, errors.ErrTokenMissing
	}

	token, ok := tokenVal.(*oauth2.Token)
	if !ok {
		return nil, errors.ErrTokenMalformed
	}

	return token, nil
}

func SessionType(session *sessions.Session) (types.SessionType, error) {
	sessionTypeVal := session.Values[sessionTypeKey]
	if sessionTypeVal == nil {
		return "", errors.ErrUserTypeMissing
	}
	sessionType, ok := sessionTypeVal.(types.SessionType)
	if !ok {
		return "", errors.ErrUserTypeMalformed
	}
	return sessionType, nil
}

func Origin(session *sessions.Session) (string, error) {
	originVal := session.Values[originKey]
	if originVal == nil {
		return "", errors.ErrOriginMissing
	}
	origin, ok := originVal.(string)
	if !ok {
		return "", errors.ErrOriginMalformed
	}
	return origin, nil
}

func Identifier(session *sessions.Session) (string, error) {
	identifierVal := session.Values[identifierKey]
	if identifierVal == nil {
		return "", errors.ErrIdentifierMissing
	}
	identifier, ok := identifierVal.(string)
	if !ok {
		return "", errors.ErrIdentifierMalformed
	}
	return identifier, nil
}

func SetJamLabel(session *sessions.Session, jamLabel string) {
	session.Values[jamLabelKey] = jamLabel
}

func SetToken(session *sessions.Session, token *oauth2.Token) {
	session.Values[tokenKey] = token
}

func SetSessionType(session *sessions.Session, userType types.SessionType) {
	session.Values[sessionTypeKey] = userType
}

func SetOrigin(session *sessions.Session, origin string) {
	session.Values[originKey] = origin
}

func SetIdentifier(session *sessions.Session, identifier string) {
	session.Values[identifierKey] = identifier
}
