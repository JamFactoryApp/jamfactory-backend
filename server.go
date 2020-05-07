package main

import (
	"context"
	"github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jamfactory-backend/controller"
	"jamfactory-backend/models"
	"log"
	"net/http"
	"os"
	"time"
)



func main() {

	// Load ENV variables
	enverr := godotenv.Load()
	if enverr != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println("[INFO] Loaded environment...")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_DB")))
	if err != nil {
		log.Panic("[ERROR] Error in connecting to database ", err)
	}
	log.Println("[INFO] Connected to database...")

	ctx, _ = context.WithTimeout(context.Background(), 30*time.Second)
	db.Database("jamfactory").Collection("Sessions").Drop(ctx)

	socket, err := socketio.NewServer(nil)


	env := &models.Env{
		DB: models.DB{Database: db.Database("jamfactory")},
		Store: models.NewSessionStore(db.Database("jamfactory").Collection("Sessions"), 3600, []byte("keybordcat")),
		PartyController: &models.PartyController{Socket: socket},
		}


	router := mux.NewRouter()

	authRouter := router.PathPrefix("/auth").Subrouter()
	partyRouter := router.PathPrefix("/party").Subrouter()
	queueRouter := router.PathPrefix("/queue").Subrouter()
	spotifyRouter := router.PathPrefix("/spotify").Subrouter()

	controller.RegisterAuthRoutes(authRouter, env)
	controller.RegisterPartyRoutes(partyRouter, env)
	controller.RegisterQueueRoutes(queueRouter, env)
	controller.RegisterSpotifyRoutes(spotifyRouter, env)

	controller.RegisterSocketRoutes(socket, env)

	go socket.Serve()
	defer socket.Close()

	http.Handle("/socket.io/", socket)
	http.Handle("/", router)

	log.Println("[INFO] Registered routes...")

	log.Println("[INFO] Listening on Port 3000...")
	serverErr := http.ListenAndServe(":3000", nil)

	if serverErr != nil {
		log.Fatalln(serverErr)
	}

}
