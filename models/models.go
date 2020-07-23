package models

import (
	log "github.com/sirupsen/logrus"
)

func Setup() {
	initRedisClient()
	log.Info("Initialized redis client")

	initSessionStore()
	log.Info("Initialized session store")
}
