package main

import (
	"crypto/tls"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"

	"github.com/jamfactoryapp/jamfactory-backend/api/server"
	"github.com/jamfactoryapp/jamfactory-backend/api/users"
	pkgredis "github.com/jamfactoryapp/jamfactory-backend/internal/redis"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/cache"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/config"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamfactory"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	log.SetLevel(log.TraceLevel)
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})

	if err := godotenv.Load(); err != nil {
		log.Warn(err)
	}

	conf := config.New()

	// Create JAM_DATA_DIR
	if _, err := os.Stat(conf.DataDir); os.IsNotExist(err) {
		if err := os.Mkdir(conf.DataDir, 0700); err != nil {
			log.Fatal("JAM_DATA_DIR could not be created: ", err)
		}
	}

	// Create redis pool
	pool, err := pkgredis.NewPool(conf.RedisAddress, conf.RedisPassword, conf.RedisDatabase)
	if err != nil {
		log.Fatal("could not connect to redis: ", err)
	}
	log.Debug("Initialized connection to redis")

	// Create redis stores
	redisStore := sessions.NewRedisSessionStore(pool, path.Join(conf.DataDir, ".keypairs"), conf.CookieSameSite, conf.CookieSecure)
	userStore := users.NewRedisUserStore(pool)
	log.Debug("Initialized redis stores")

	// Create redis cache
	redisCache := cache.NewRedis(pool)
	log.Debug("Initialized redis cache")

	// Create JamFactory
	spotifyJamFactory := jamfactory.NewSpotify(redisCache, conf.SpotifyRedirectURL, conf.SpotifyID, conf.SpotifySecret, conf.ClientAddress.String())
	log.Debug("Initialized JamFactory")

	// Create app server
	appServer := server.NewServer("/", redisStore, userStore, spotifyJamFactory).
		WithPort(conf.Port).
		WithCache(redisCache)

	if conf.UseHttps {
		// Optionally create self-signed certificates for HTTPS
		if conf.GenCerts {
			server.GenCerts(conf.DNSNames, conf.CertFile, conf.KeyFile)
		}

		appServer = appServer.WithTLS(&tls.Config{
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
		})
		log.Infof("HTTPS server is listening on :%d\n", conf.Port)
		if err := appServer.RunTLS(conf.CertFile, conf.KeyFile); err != nil {
			log.Fatal("HTTPS server failed to listen: ", err)
		}
	} else {
		log.Infof("HTTP server is listening on :%d\n", conf.Port)
		if err := appServer.Run(); err != nil {
			log.Fatal("HTTP server failed to listen: ", err)
		}
	}
}
