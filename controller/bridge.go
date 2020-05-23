package controller

import (
	socketio "github.com/googollee/go-socket.io"
	"jamfactory-backend/models"
	"os"
)

var Store *SessionStore
var Socket *socketio.Server
var Factory models.Factory

func Setup() {
	db := models.MongoClient.Database(os.Getenv("MONGO_DB_NAME"))
	collection := db.Collection(models.MongoSessions)
	Store = NewSessionStore(collection, 3600, []byte("keybordcat"))

	Factory = models.Factory{
		Partys: nil,
		Count:  0,
		Socket: Socket,
	}
}
