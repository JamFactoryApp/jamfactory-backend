package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"strings"
)

func JamClientAddress() string {
	address := os.Getenv("JAM_CLIENT_ADDRESS")
	if !strings.HasPrefix(address, "http") {
		log.Fatal("JAM_CLIENT_ADDRESS requires protocol scheme (e.g. http://)")
	}
	port, err := strconv.Atoi(os.Getenv("JAM_CLIENT_PORT"))
	if err != nil {
		log.Fatal("Invalid client port")
	}
	return fmt.Sprintf("%s:%d", address, port)
}
