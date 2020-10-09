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

func JamClient() string {
	address := os.Getenv("JAM_CLIENT_ADDRESS")
	if !strings.HasPrefix(address, "http") {
		log.Fatal("JAM_CLIENT_ADDRESS requires protocol scheme (e.g. http://)")
	}
	if strings.HasPrefix(address, "http://") {
		return address[7:]
	} else if strings.HasPrefix(address, "https://") {
		return address[8:]
	} else {
		log.Fatal("Invalid protocol scheme for JAM_CLIENT_ADDRESS")
	}
	return ""
}
