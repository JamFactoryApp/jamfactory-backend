package config

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	APIAddress         *url.URL
	ClientAddress      *url.URL
	CookieSameSite     http.SameSite
	CookieSecure       bool
	DataDir            string
	LogLevel           log.Level
	RedisAddress       string
	RedisPassword      string
	RedisDatabase      string
	SpotifyRedirectURL string
	SpotifyID          string
	SpotifySecret      string
}

func New() *Config {
	var err error
	c := &Config{}

	apiAddress := os.Getenv("JAM_API_ADDRESS")
	if apiAddress == "" {
		log.Fatal("JAM_API_ADDRESS is empty")
	}
	c.APIAddress, err = url.Parse(apiAddress)
	if err != nil {
		log.Fatal("failed to parse JAM_API_ADDRESS: ", err)
	}

	clientAddress := os.Getenv("JAM_CLIENT_ADDRESS")
	if clientAddress == "" {
		log.Fatal("JAM_CLIENT_ADDRESS is empty")
	}
	c.ClientAddress, err = url.Parse(clientAddress)
	if err != nil {
		log.Fatal("failed to parse JAM_CLIENT_ADDRESS: ", err)
	}

	c.DataDir = os.Getenv("JAM_DATA_DIR")

	var logLevel log.Level
	parsedLogLevel, err := log.ParseLevel(os.Getenv("JAM_LOG_LEVEL"))
	if parsedLogLevel == logLevel {
		logLevel = log.WarnLevel
	}
	if err == nil {
		logLevel = parsedLogLevel
	}
	c.LogLevel = logLevel

	c.RedisAddress = os.Getenv("JAM_REDIS_ADDRESS")

	c.RedisPassword = os.Getenv("JAM_REDIS_PASSWORD")
	c.RedisDatabase = os.Getenv("JAM_REDIS_DATABASE")

	c.SpotifyID = os.Getenv("JAM_SPOTIFY_ID")
	if c.SpotifyID == "" {
		log.Fatal("JAM_SPOTIFY_ID cannot be empty")
	}
	c.SpotifySecret = os.Getenv("JAM_SPOTIFY_SECRET")
	if c.SpotifySecret == "" {
		log.Fatal("JAM_SPOTIFY_SECRET cannot be empty")
	}
	c.SpotifyRedirectURL = os.Getenv("JAM_SPOTIFY_REDIRECT_URL")
	if c.SpotifyRedirectURL == "" {
		log.Fatal("JAM_SPOTIFY_REDIRECT_URL cannot be empty")
	}

	environment := os.Getenv("JAM_PRODUCTION")

	switch strings.ToLower(environment) {
	case "production":
		c.CookieSecure = true
	default:
		c.CookieSecure = false
	}
	c.CookieSameSite = http.SameSiteLaxMode

	return c
}
