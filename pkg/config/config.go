package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	APIAddress         string
	CookieSameSite     http.SameSite
	CookieSecure       bool
	LogLevel           log.Level
	RedisAddress       string
	RedisPort          int
	RedisPassword      string
	RedisDatabase      string
	SpotifyRedirectURL string
	SpotifyID          string
	SpotifySecret      string
}

func New() *Config {
	c := &Config{}

	apiAddress := strings.TrimPrefix(strings.TrimPrefix(os.Getenv("JAM_API_ADDRESS"), "https://"), "http://")
	apiPort, err := strconv.Atoi(os.Getenv("JAM_API_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	if apiPort == 80 {
		c.APIAddress = apiAddress
	} else {
		c.APIAddress = fmt.Sprintf("%s:%d", apiAddress, apiPort)
	}

	var logLevel log.Level
	ll, err := log.ParseLevel(os.Getenv("JAM_LOG_LEVEL"))
	if err == nil {
		logLevel = ll
	}
	c.LogLevel = logLevel

	c.RedisAddress = os.Getenv("JAM_REDIS_ADDRESS")

	redisPort, err := strconv.Atoi(os.Getenv("JAM_REDIS_PORT"))
	if err != nil {
		log.Fatal(err)
	}
	c.RedisPort = redisPort

	c.RedisPassword = os.Getenv("JAM_REDIS_PASSWORD")
	c.RedisDatabase = os.Getenv("JAM_REDIS_DATABASE")
	c.SpotifyRedirectURL = os.Getenv("JAM_SPOTIFY_REDIRECT_URL")
	c.SpotifyID = os.Getenv("JAM_SPOTIFY_ID")
	c.SpotifySecret = os.Getenv("JAM_SPOTIFY_SECRET")

	environment := os.Getenv("JAM_PRODUCTION")

	switch strings.ToLower(environment) {
	case "production":
		c.CookieSameSite = http.SameSiteLaxMode
		c.CookieSecure = true
	default:
		c.CookieSameSite = http.SameSiteNoneMode
		c.CookieSecure = false
	}

	return c
}
