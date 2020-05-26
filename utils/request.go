package utils

import (
	"github.com/gorilla/sessions"
	"jamfactory-backend/models"
	"net/http"
)

const (
	JamSessionContextKey = "JamSession"
	SessionContextKey    = "Session"
)

func JamSessionFromRequestContext(r *http.Request) *models.JamSession {
	return r.Context().Value(JamSessionContextKey).(*models.JamSession)
}

func SessionFromRequestContext(r *http.Request) *sessions.Session {
	return r.Context().Value(SessionContextKey).(*sessions.Session)
}
