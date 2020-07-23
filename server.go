package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/controllers"
	"jamfactory-backend/models"
	"math/rand"
	"net/http"
	"os"
	"time"
)

const (
	port         = 3000
	readTimeout  = time.Second
	writeTimeout = time.Second
	idleTimeout  = time.Second
)

var (
	server      *http.Server
	corsOptions = cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:4200"},
		AllowCredentials: true,
		Debug:            false,
	}
	//corsAll := cors.Default().Handler(controllers.Router)
)

func setup() {
	rand.Seed(time.Now().UTC().UnixNano())

	initLogging()

	initEnvironment()
	log.Info("Initialized environment")

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
}

func initHttpServer() {
	corsHandler := cors.New(corsOptions).Handler(controllers.Router)
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
