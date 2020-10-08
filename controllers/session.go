package controllers

import (
	"github.com/gorilla/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
)

const (
	storeMaxAge = 60 * 60 * 24 * 7
)

var (
	store         *utils.RedisStore
	storeRedisKey = utils.RedisKey{}.Append("session")
)

func initSessionStore() {
	conn := models.RedisPool.Get()
	if conn.Err() != nil {
		log.Fatal("Connection to redis could not be established!")
	}
	cookieKeyPairsCount, err := strconv.Atoi(os.Getenv("JAM_COOKIE_KEY_PAIRS_COUNT"))
	if err != nil {
		log.Fatal(err)
	}
	store = utils.NewRedisStore(conn, storeRedisKey, storeMaxAge, cookieKeyPairsCount)
}

func GetSession(r *http.Request, name string) (*sessions.Session, error) {
	return store.Get(r, name)
}

func SaveSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) {
	err := session.Save(r, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Session", session.ID).Warnf("Could not save session: %s\n", err.Error())
	}
}

func SessionIsValid(session *sessions.Session) bool {
	return session.Values[utils.SessionUserTypeKey] != nil
}

func LoggedInAsHost(session *sessions.Session) bool {
	return SessionIsValid(session) && session.Values[utils.SessionUserTypeKey] == models.UserTypeHost
}

func LoggedIntoSpotify(session *sessions.Session) (bool, error) {
	token, err := utils.ParseTokenFromSession(session)
	if err != nil {
		return false, err
	}
	return SessionIsValid(session) && session.Values[utils.SessionTokenKey] != nil && token.Valid(), nil
}
