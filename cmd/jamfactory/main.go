package main

import (
	"fmt"
	"github.com/jamfactoryapp/jamfactory-backend/api/server"
	"github.com/jamfactoryapp/jamfactory-backend/api/store"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/config"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamfactory"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"path"
	"time"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	rand.Seed(time.Now().UnixNano())

	log.SetLevel(log.WarnLevel)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})

	if err := godotenv.Load(); err != nil {
		log.Warn(err)
	}

	conf := config.New()

	log.SetLevel(conf.LogLevel)

	if _, err := os.Stat(conf.DataDir); os.IsNotExist(err) {
		if err := os.Mkdir(conf.DataDir, 0700); err != nil {
			return err
		}
	}

	pool, err := pkgredis.NewPool(conf.RedisAddress, conf.RedisPort, conf.RedisPassword, conf.RedisDatabase)
	if err != nil {
		return err
	}
	log.Debug("Initialized connection to redis")

	st := store.NewRedis(pool, path.Join(conf.DataDir, ".keypairs"), conf.CookieSameSite, conf.CookieSecure)
	log.Debug("Initialized redis cookie store")

	ca := cache.NewRedis(pool)
	log.Debug("Initialized redis cache")

	ja := jamfactory.NewSpotify(ca, conf.SpotifyRedirectURL, conf.SpotifyID, conf.SpotifySecret)
	log.Debug("Initialized JamFactory")

	se := server.NewServer("/", st, ja).
		WithAddress(conf.APIAddress).
		WithCache(ca)

	log.Infof("HTTP server is listening on %s\n", conf.APIAddress)
	return se.Run()
}
