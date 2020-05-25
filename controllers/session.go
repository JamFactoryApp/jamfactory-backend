package controllers

import (
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"net/http"
)

const (
	storeMaxAge = 3600

	SessionUserTypeKey = "User"
	SessionLabelKey    = "Label"
	SessionTokenKey    = "Token"
)

var (
	Store         *models.SessionStore
	storeKeyPairs = []byte("keyboardcat")
)

func initSessionStore() {
	Store = models.NewSessionStore(storeMaxAge, storeKeyPairs)
}

func GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return Store.Get(r, name)
}

func SaveSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	err := session.Save(r, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Session", session.ID).Warnf("Could not save session:\n%s\n", err.Error())
	}
}

func SessionIsValid(session *sessions.Session) bool {
	return session.Values[SessionUserTypeKey] != nil && session.Values[SessionTokenKey] != nil
}

func LoggedInAsHost(session *sessions.Session) bool {
	return SessionIsValid(session) && session.Values[SessionUserTypeKey] == models.UserTypeHost
}
