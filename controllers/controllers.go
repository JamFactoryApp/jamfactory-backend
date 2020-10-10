package controllers

import (
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
)

var (
	afterCallbackRedirect string
)

func Setup() {
	initVars()
	log.Info("Initialized vars")

	initSessionStore()
	log.Info("Initialized session store")

	initSpotifyAuthenticator()
	log.Info("Initialized Spotify authenticator")

	initMiddleWares()
	log.Info("Initialized middlewares")

	initRoutes()
	log.Info("Initialized routes")

	initFactory()
	log.Info("Initialized factory")
}

func initVars() {
	afterCallbackRedirect = utils.JamClientAddress()
}
