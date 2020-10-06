package utils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
)

func JamClientAddress() string {
	address := os.Getenv("JAM_CLIENT_ADDRESS")
	port, err := strconv.Atoi(os.Getenv("JAM_CLIENT_PORT"))
	if err != nil {
		log.Fatal("Invalid client port")
	}
	return fmt.Sprintf("%s:%d", address, port)
}
