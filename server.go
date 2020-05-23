package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/controller"
	"jamfactory-backend/models"
	"net/http"
	"os"
	"time"
)

var PORT = 3000

// Load ENV variables
func loadEnvironment() {
	err := godotenv.Load()

	if err != nil {
		log.Error("No .env file found", err)
	}
}

func main() {

	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: false,
	})
	log.SetLevel(log.TraceLevel)

	loadEnvironment()
	log.Info("Loaded environment")

	models.InitDB()
	log.Info("Initialized database")

	controller.Setup()
	log.Info("Initialized controllers")

	controller.SpotifyAuthenticator = spotify.NewAuthenticator(
		os.Getenv("SPOTIFY_REDIRECT_URL"),
		spotify.ScopeUserReadPrivate,
		spotify.ScopeUserReadEmail,
		spotify.ScopeUserModifyPlaybackState,
		spotify.ScopeUserReadPlaybackState,
	)

	router := mux.NewRouter()

	authRouter := router.PathPrefix("/api/auth").Subrouter()
	partyRouter := router.PathPrefix("/api/party").Subrouter()
	queueRouter := router.PathPrefix("/api/queue").Subrouter()
	spotifyRouter := router.PathPrefix("/api/spotify").Subrouter()


	controller.RegisterAuthRoutes(authRouter)
	controller.RegisterPartyRoutes(partyRouter)
	controller.RegisterQueueRoutes(queueRouter)
	controller.RegisterSpotifyRoutes(spotifyRouter)
	log.Info("Initialized routes")

	socket := controller.InitSocketIO()

	go socket.Serve()
	defer socket.Close()
	controller.Socket = socket
	controller.Factory.SetSocket(socket)
	log.Info("Initialized socketio server")
	socketRouter := router.PathPrefix("/socket.io/").Subrouter()
	socketRouter.Handle("/", socket)

	http.Handle("/", router)

	go queueWorker(&controller.Factory)

	log.Infof("Listening on Port %v", PORT)

	corsOptions := cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:4200"},
		AllowCredentials: true,
		Debug:            false,
	}

	corsHandler := cors.New(corsOptions).Handler(router)
	//corsAll := cors.Default().Handler(router)



	err := http.ListenAndServe(fmt.Sprintf(":%v", PORT), corsHandler)

	if err != nil {
		log.Fatal(err)
	}

}

func queueWorker(partyController *models.Factory) {
	for {
		time.Sleep(1 * time.Second)
		go models.QueueWorker(partyController)
	}
}

