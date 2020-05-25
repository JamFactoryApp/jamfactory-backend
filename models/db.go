package models

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

const connectTimeout = 3 * time.Second

var (
	db          *mongo.Database
	mongoClient *mongo.Client
)

func initMongoClient() {
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	opts := options.Client().ApplyURI(os.Getenv("MONGO_DB"))

	var err error
	mongoClient, err = mongo.Connect(ctx, opts)

	if err != nil {
		log.WithContext(ctx).Panic("Error connecting to database: ", err.Error())
	}
}

func initDb() {
	db = mongoClient.Database(os.Getenv("MONGO_DB_NAME"))
}
