package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"jamfactory-backend/controller"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()

	authRouter := router.PathPrefix("/auth").Subrouter()
	partyRouter := router.PathPrefix("/party").Subrouter()
	queueRouter := router.PathPrefix("/queue").Subrouter()
	spotifyRouter := router.PathPrefix("/spotify").Subrouter()

	controller.RegisterAuthRoutes(authRouter)
	controller.RegisterPartyRoutes(partyRouter)
	controller.RegisterQueueRoutes(queueRouter)
	controller.RegisterSpotifyRoutes(spotifyRouter)

	http.Handle("/", router)

	fmt.Println("Listening on Port 3000....")
	err := http.ListenAndServe(":3000", nil)

	if err != nil {
		log.Fatalln(err)
	}
}
