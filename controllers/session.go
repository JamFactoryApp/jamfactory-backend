package controllers

import (
	"errors"
	"github.com/gorilla/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	storeMaxAge            = 60 * 60 * 24 * 7
	redisConnRetries       = 3
	redisConnRetryInterval = 2 * time.Second
)

var (
	store         *utils.RedisStore
	storeRedisKey = utils.RedisKey{}.Append("session")
)

func initSessionStore() {
	err := tryRedisConn()
	if err != nil {
		log.Fatal("Failed to connect to redis: ", err)
	}

	cookieKeyPairsCount, err := strconv.Atoi(os.Getenv("JAM_COOKIE_KEY_PAIRS_COUNT"))
	if err != nil {
		log.Fatal(err)
	}
	store = utils.NewRedisStore(models.RedisPool, storeRedisKey, storeMaxAge, cookieKeyPairsCount)
}

func tryRedisConn() error {
	for i := 0; i < redisConnRetries; i++ {
		conn := models.RedisPool.Get()
		if conn.Err() != nil {
			log.Warn("Connection to redis could not be established: ", conn.Err(), " Retrying in ", redisConnRetryInterval, " seconds")
			time.Sleep(redisConnRetryInterval)
		} else {
			return nil
		}
	}
	return errors.New("max retries exceeded")
}

func GetSession(r *http.Request, name string) *sessions.Session {
	session, _ := store.Get(r, name)
	return session
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
