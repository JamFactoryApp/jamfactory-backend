package sessions

import (
	"context"
	"github.com/gorilla/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/errors"
)

type contextKey string

const key contextKey = "Session"

const (
	identifierKey = "Identifier"
	originKey     = "Origin"
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

func SetOrigin(session *sessions.Session, origin string) {
	session.Values[originKey] = origin
}

func SetIdentifier(session *sessions.Session, identifier string) {
	session.Values[identifierKey] = identifier
}
