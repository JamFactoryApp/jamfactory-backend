package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/controllers"
	"jamfactory-backend/models"
	"jamfactory-backend/types"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	port         = 3000
	readTimeout  = time.Second
	writeTimeout = time.Second
	idleTimeout  = time.Second
)

var (
	server          *http.Server
	requiredEnvVars = []string{
		"ALLOWED_CLIENTS",
		"ALLOWED_CLIENTS_SEP",
		"ALLOWED_HEADERS",
		"ALLOWED_HEADERS_SEP",
		"SPOTIFY_ID",
		"SPOTIFY_SECRET",
		"SPOTIFY_REDIRECT_URL",
		"REDIS_ADDRESS",
		"REDIS_DATABASE",
		"REDIS_PASSWORD",
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
	log.SetLevel(log.TraceLevel)
}

func initEnvironment() {
	if err := godotenv.Load(); err != nil {
		log.Warnf("No .env.example file found:\n%s\n", err)
	}

	var notDefined []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			notDefined = append(notDefined, envVar)
		}
	}
	if len(notDefined) > 0 {
		log.Fatalf("The following environment variables are not defined: %v\n", notDefined)
	}
}

func initHttpServer() {
	allowedOrigins := handlers.AllowedOrigins(strings.Split(os.Getenv("ALLOWED_CLIENTS"), os.Getenv("ALLOWED_CLIENTS_SEP")))
	allowedHeaders := handlers.AllowedHeaders(strings.Split(os.Getenv("ALLOWED_HEADERS"), os.Getenv("ALLOWED_HEADERS_SEP")))
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "OPTIONS"})
	allowCredentials := handlers.AllowCredentials()
	corsHandler := handlers.CORS(allowedOrigins, allowedHeaders, allowedMethods, allowCredentials)(controllers.Router)

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

	log.Infof("HTTP server is listening on port %v\n", port)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Error while listening and serving:\n%s\n", err)
	}
}
