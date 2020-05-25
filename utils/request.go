package utils

import (
	"github.com/gorilla/sessions"
	"jamfactory-backend/models"
	"net/http"
)

const (
	PartyContextKey   = "Party"
	SessionContextKey = "Session"
)

func PartyFromRequestContext(r *http.Request) *models.Party {
	return r.Context().Value(PartyContextKey).(*models.Party)
}

func SessionFromRequestContext(r *http.Request) *sessions.Session {
	return r.Context().Value(SessionContextKey).(*sessions.Session)
}
