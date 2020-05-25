package controllers

import (
	log "github.com/sirupsen/logrus"
)

func Setup() {
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

	initSessionStore()
	log.Info("Initialized session store")
}
