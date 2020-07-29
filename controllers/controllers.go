package controllers

import (
	log "github.com/sirupsen/logrus"
)

func Setup() {
	initSessionStore()
	log.Info("Initialized session store")

	initSpotifyAuthenticator()
	log.Info("Initialized Spotify authenticator")

	initSocketIO()
	log.Info("Initialized socket.io server")

	initMiddleWares()
	log.Info("Initialized middlewares")

	initRoutes()
	log.Info("Initialized routes")

	initFactory()
	log.Info("Initialized factory")
}
