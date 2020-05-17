package models

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

const ConnectTimeout = 3 * time.Second
const DropTimeout = 3 * time.Second
const MongoSessions = "Sessions"

var MongoClient *mongo.Client

// Try connecting to database
func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), ConnectTimeout)
	defer cancel()

	opts := options.Client().ApplyURI(os.Getenv("MONGO_DB"))

	var err error
	MongoClient, err = mongo.Connect(ctx, opts)

	if err != nil {
		log.WithContext(ctx).Panic("Error connecting to database: ", err.Error())
	}

	DropOldSessions()
}

// Try dropping old session data from database
func DropOldSessions() {
	ctx, cancel := context.WithTimeout(context.Background(), DropTimeout)
	defer cancel()

	sessions := MongoClient.Database(os.Getenv("MONGO_DB_NAME"), nil).Collection(MongoSessions, nil)
	err := sessions.Drop(ctx)

	if err != nil {
		log.WithContext(ctx).Panic("Error dropping old session data: ", err.Error())
	} else {
		log.Warn("Dropped old session data")
	}

}
