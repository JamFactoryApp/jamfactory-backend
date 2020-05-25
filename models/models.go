package models

import (
	log "github.com/sirupsen/logrus"
)

func Setup() {
	initMongoClient()
	log.Info("Initialized MongoDB client")

	initDb()
	log.Info("Initialized database")

	initSessionsCollection()
	log.Info("Initialized sessions collection")

	dropOldSessions()
	log.Warn("Dropped old sessions")
}
