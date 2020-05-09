package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/zmb3/spotify"
	"jamfactory-backend/controller"
	"jamfactory-backend/models"
	"log"
	"net/http"
	"os"
	"time"
)

var PORT = 3000

// Load ENV variables
func loadEnvironment() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found\n", err)
	}
}

func main() {
	loadEnvironment()
	log.Println("Loaded environment")

	models.InitDB()
	log.Println("Initialized database")

	controller.Setup()
	log.Println("Initialized controllers")

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
	log.Println("Initialized routes")

	socket := controller.InitSocketIO()
	go socket.Serve()
	defer socket.Close()
	controller.Socket = socket
	log.Println("Initialized socketio server")

	http.Handle("/", router)
	http.Handle("/socket.io/", socket)

	go queueWorker(&controller.PartyControl)

	log.Printf("Listening on Port %v\n", PORT)
	err := http.ListenAndServe(fmt.Sprintf(":%v", PORT), handlers.CORS()(router))

	if err != nil {
		log.Fatalln(err)
	}

}

func queueWorker(partyController *controller.PartyController) {
	for {
		time.Sleep(1 * time.Second)
		go controller.QueueWorker(partyController)
	}
}
