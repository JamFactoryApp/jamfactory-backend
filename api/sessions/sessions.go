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
	jamLabelKey = "Label"
	tokenKey    = "CurrentToken"
	userTypeKey = "User"
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

func UserType(session *sessions.Session) (types.UserType, error) {
	userTypeVal := session.Values[userTypeKey]
	if userTypeVal == nil {
		return "", errors.ErrUserTypeMissing
	}
	userType, ok := userTypeVal.(types.UserType)
	if !ok {
		return "", errors.ErrUserTypeMalformed
	}
	return userType, nil
}

func SetJamLabel(session *sessions.Session, jamLabel string) {
	session.Values[jamLabelKey] = jamLabel
}

func SetToken(session *sessions.Session, token *oauth2.Token) {
	session.Values[tokenKey] = token
}

func SetUserType(session *sessions.Session, userType types.UserType) {
	session.Values[userTypeKey] = userType
}
