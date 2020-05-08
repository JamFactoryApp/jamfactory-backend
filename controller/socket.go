package controller

import (
	"errors"
	socketio "github.com/googollee/go-socket.io"
	"log"
	"net/http"
)

// Initialize socketio server
func InitSocketIO() *socketio.Server {
	socket, err := socketio.NewServer(nil)

	if err != nil {
		log.Fatalln("Error creating socketio server\n", err)
	}

	RegisterSocketRoutes(socket)

	return socket
}

func RegisterSocketRoutes(socket *socketio.Server) {
	socket.OnConnect("/", SocketAuth)
	socket.OnEvent("/", "connection", SocketConnect)
}

func SocketConnect(s socketio.Conn, msg string) {
	session, _ := Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")

	s.Join(session.Values["label"].(string))
}

func SocketAuth(s socketio.Conn) error {
	session, _ := Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")
	if session.Values["user"] == "Host" || session.Values["user"] == "Guest" {
		if PartyControl.GetParty(session.Values["label"].(string)) != nil {
			return nil
		} else {
			return errors.New("label is invalid")
		}
	}
	return errors.New("wrong user type")
}
