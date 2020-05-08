package models

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

const ConnectTimeout = 3 * time.Second
const DropTimeout = 3 * time.Second

var MongoClient *mongo.Client

// Try connecting to database
func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer cancel()

	opts := options.Client().ApplyURI(os.Getenv("MONGO_DB"))

	var err error
	MongoClient, err = mongo.Connect(ctx, opts)

	if err != nil {
		log.Fatalln("Error connecting to database\n", err)
	}

	DropOldSessions()
	log.Println("Dropped old session data")
}

// Try dropping old session data from database
func DropOldSessions() {
	ctx, cancel := context.WithTimeout(context.Background(), DropTimeout)
	defer cancel()

	sessions := MongoClient.Database(os.Getenv("MONGO_DB_NAME"), nil).Collection(os.Getenv("MONGO_DB_SESSIONS"), nil)
	err := sessions.Drop(ctx)

	if err != nil {
		log.Fatalln("Error dropping old session data\n", err)
	}
}
