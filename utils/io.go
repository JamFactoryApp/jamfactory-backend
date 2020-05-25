package utils

import (
	log "github.com/sirupsen/logrus"
	"io"
)

func CloseProperly(closeable io.Closer) {
	err := closeable.Close()
	if err != nil {
		log.Panic("Error colsing socket.io server")
	}
}
