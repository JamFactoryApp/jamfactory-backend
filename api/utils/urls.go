package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
)

func JamRedirectURL() string {
	var once sync.Once
	var clientAddress string
	var clientURL string
	once.Do(func() {
		clientURL = os.Getenv("JAM_CLIENT_ADDRESS")
		clientPortStr := os.Getenv("JAM_CLIENT_PORT")
		p, err := strconv.Atoi(clientPortStr)
		if err != nil {
			log.Fatal(err)
		}
		if p == 80 {
			clientAddress = clientURL
		} else {
			clientAddress = fmt.Sprintf("%s:%d", clientURL, p)
		}
	})
	return clientAddress
}
