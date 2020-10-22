package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/jamfactoryapp/jamfactory-backend/controllers"
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/types"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	readTimeout  = time.Second
	writeTimeout = time.Second
	idleTimeout  = time.Second
)

var (
	server          *http.Server
	port            int
	production		bool
	requiredEnvVars = []string{
		"JAM_API_ADDRESS",
		"JAM_PRODUCTION",
		"JAM_API_PORT",
		"JAM_CLIENT_ADDRESS",
		"JAM_CLIENT_PORT",
		"JAM_SPOTIFY_ID",
		"JAM_SPOTIFY_SECRET",
		"JAM_SPOTIFY_REDIRECT_URL",
		"JAM_REDIS_ADDRESS",
		"JAM_REDIS_PORT",
		"JAM_REDIS_DATABASE",
	}
)

func setup() {
	rand.Seed(time.Now().UTC().UnixNano())

	initLogging()

	initEnvironment()
	log.Info("Initialized environment")

	types.RegisterGobTypes()
	log.Info("Initialized types")

	models.Setup()
	log.Info("Initialized models")

	controllers.Setup()
	log.Info("Initialized controllers")

	initHttpServer()
	log.Info("Initialized HTTP server")
}

func initLogging() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})
}

func initEnvironment() {
	if err := godotenv.Load(); err != nil {
		log.Warn(err)
	}

	var notDefined []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			notDefined = append(notDefined, envVar)
		}
	}
	if len(notDefined) > 0 {
		log.Fatal("The following environment variables are not defined: ", notDefined)
	}

	prod, ok := os.LookupEnv("JAM_PRODUCTION")
	if ok && strings.ToLower(prod) == "production" {
		production = true
	} else {
		production = false
		log.Warn("You are running JamFactory in the development environment")
	}

	initLogLevel()
}

func initLogLevel() {
	var logLevel log.Level

	level, ok := os.LookupEnv("JAM_LOG_LEVEL")
	if !ok {
		level = "INFO"
	}

	switch strings.ToLower(level) {
	case "panic":
		logLevel = log.PanicLevel
	case "fatal":
		logLevel = log.FatalLevel
	case "error":
		logLevel = log.ErrorLevel
	case "warn":
		logLevel = log.WarnLevel
	case "info":
		logLevel = log.InfoLevel
	case "debug":
		logLevel = log.DebugLevel
	case "trace":
		logLevel = log.TraceLevel
	default:
		log.Fatal("Invalid log level")
	}

	log.SetLevel(logLevel)
}

func initHttpServer() {
	allowedOrigins := handlers.AllowedOrigins([]string{utils.JamClientAddress()})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "OPTIONS"})
	allowCredentials := handlers.AllowCredentials()
	corsHandler := handlers.CORS(allowedOrigins, allowedHeaders, allowedMethods, allowCredentials)(controllers.Router)

	var err error
	port, err = strconv.Atoi(os.Getenv("JAM_API_PORT"))
	if err != nil {
		log.Fatal("Invalid api port")
	}

	server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		Handler:      corsHandler,
	}
	http.Handle("/", controllers.Router)
}

func main() {
	setup()
	log.Info("Setup complete")

	log.Info("HTTP server is listening on port ", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Error while listening and serving: ", err)
	}
}
