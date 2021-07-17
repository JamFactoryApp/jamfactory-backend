package ioutils

import (
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

func CloseProperly(closeable io.Closer) {
	err := closeable.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
