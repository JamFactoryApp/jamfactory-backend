package utils

import log "github.com/sirupsen/logrus"

func Setup() {
	registerGobTypes()
	log.Info("Registered gob types")
}
