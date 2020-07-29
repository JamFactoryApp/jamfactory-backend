package controllers

import (
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
	"net/http"
)

const (
	storeMaxAge = 3600
)

var (
	Store         *models.RedisStore
	storeKeyPairs = []byte("keyboardcat")
)

func initSessionStore() {
	conn := models.RedisPool.Get()
	Store = models.NewRedisStore(conn, "session:", storeMaxAge, storeKeyPairs)
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
	return session.Values[models.SessionUserTypeKey] != nil
}

func LoggedInAsHost(session *sessions.Session) bool {
	return SessionIsValid(session) && session.Values[models.SessionUserTypeKey] == models.UserTypeHost
}

func LoggedIntoSpotify(session *sessions.Session) (bool, error) {
	token, err := utils.ParseTokenFromSession(session)
	if err != nil {
		return false, err
	}
	return SessionIsValid(session) && session.Values[models.SessionTokenKey] != nil && token.Valid(), nil
}
