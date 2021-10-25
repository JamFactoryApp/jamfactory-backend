package utils

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
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

func SplitsIds(arr []spotify.ID, chunkSize int) [][]spotify.ID {
	if len(arr) == 0 {
		return nil
	}
	divided := make([][]spotify.ID, (len(arr)+chunkSize-1)/chunkSize)
	prev := 0
	i := 0
	till := len(arr) - chunkSize
	for prev < till {
		next := prev + chunkSize
		divided[i] = arr[prev:next]
		prev = next
		i++
	}
	divided[i] = arr[prev:]
	return divided
}
