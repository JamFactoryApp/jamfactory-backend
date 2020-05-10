package controller

import (
	"errors"
	"fmt"
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
	socket.OnError("/", SocketError)
	socket.OnDisconnect("/", SocketDisconnect)
}

func SocketConnect(s socketio.Conn, msg string) {
	session, _ := Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")
	log.Printf("Socket.Io connection from %s", session.ID)

	log.Printf(msg)
	//s.Join(session.Values["label"].(string))
	s.Join(msg)
}

func SocketAuth(s socketio.Conn) error {

	log.Printf("Start Socket.Io auth")

	session, err := Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")

	if err != nil {
		s.Close()
		log.Printf("Could not get session")
		return err
	}

	s.LeaveAll()
	log.Printf("Socket.Io auth from %s", session.ID)
	if (session.Values["User"] == "Host" || session.Values["User"] == "Guest") && session.Values["Label"] != nil {
		if PartyControl.GetParty(session.Values["Label"].(string)) != nil {
			log.Printf("allowed")
			s.Join(session.Values["Label"].(string))
			s.SetContext(s.Context())
			return nil
		} else {
			s.Close()
			return errors.New("label is invalid")
		}
	}
	s.Close()
	return errors.New("wrong user type")
}

func SocketError(s socketio.Conn, e error) {
	fmt.Println("meet error:", e)
}

func SocketDisconnect(s socketio.Conn, reason string) {
	fmt.Println("closed", reason)
}