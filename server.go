package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jamfactory-backend/controller"
	"jamfactory-backend/models"
	"log"
	"net/http"
	"time"
)



func main() {

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Panic(err)
	}

	env := &models.Env{
		DB: models.DB{Database: db.Database("jamfactory")},
		Store: models.NewSessionStore(db.Database("jamfactory").Collection("Sessions"), 420, []byte("keybordcat")),
		Text: "Env Test"}

	router := mux.NewRouter()

	authRouter := router.PathPrefix("/auth").Subrouter()
	partyRouter := router.PathPrefix("/party").Subrouter()
	queueRouter := router.PathPrefix("/queue").Subrouter()
	spotifyRouter := router.PathPrefix("/spotify").Subrouter()

	controller.RegisterAuthRoutes(authRouter, env)
	controller.RegisterPartyRoutes(partyRouter, env)
	controller.RegisterQueueRoutes(queueRouter, env)
	controller.RegisterSpotifyRoutes(spotifyRouter, env)

	http.Handle("/", router)

	fmt.Println("Listening on Port 3000....")
	serverErr := http.ListenAndServe(":3000", nil)

	if serverErr != nil {
		log.Fatalln(serverErr)
	}
}
