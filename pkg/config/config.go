package config

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port               int
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

	portVal := os.Getenv("JAM_PORT")
	if portVal == "" {
		log.Fatal("JAM_PORT is empty")
	}
	port, err := strconv.Atoi(portVal)
	if err != nil {
		log.Fatal("failed to parse JAM_PORT: ", err)
	}
	c.Port = port

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

	environment := os.Getenv("JAM_ENV")

	switch strings.ToLower(environment) {
	case "production":
		c.CookieSecure = true
	default:
		c.CookieSecure = false
	}
	c.CookieSameSite = http.SameSiteLaxMode

	return c
}
