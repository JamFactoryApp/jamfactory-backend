package models

import (
	log "github.com/sirupsen/logrus"
)

func Setup() {
	initRedisPool()
	log.Info("Initialized redis pool")
}
