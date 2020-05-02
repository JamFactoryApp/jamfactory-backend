package models

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database interface {
	GetAllSessions() ([]*Session, error)
	SaveSession(session bson.M) (primitive.ObjectID, error)
}

type DB struct {
	*mongo.Database
}