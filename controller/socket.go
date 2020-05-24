package controller

import (
	"errors"
	socketio "github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
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
	socket.OnError("/", SocketError)
	socket.OnDisconnect("/", SocketDisconnect)
}

func SocketAuth(s socketio.Conn) error {
	session, err := Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")

	if err != nil {
		_ = s.Close()
		log.WithField("Socket", s.ID()).Warn("Could not get session")
		return err
	}

	s.LeaveAll()
	var logger = log.WithFields(log.Fields{
		"Socket":  s.ID(),
		"Session": session.ID,
	})
	logger.Trace("starting Socket.IO auth")

	if (session.Values[models.SessionUserKey] == "Host" || session.Values[models.SessionUserKey] == "Guest") && session.Values[models.SessionLabelKey] != nil {
		if Factory.GetParty(session.Values[models.SessionLabelKey].(string)) != nil {
			s.Join(session.Values[models.SessionLabelKey].(string))
			s.SetContext(s.Context())
			logger.Trace("allowed connection")
			return nil
		} else {
			_ = s.Close()
			logger.Trace("disallowed connection: label invalid")
			return errors.New("label invalid")
		}
	}
	_ = s.Close()
	logger.Trace("disallowed connection: wrong user type")
	return errors.New("wrong user type")
}

func SocketError(s socketio.Conn, e error) {
	log.Error("Socket.IO Error ", e.Error())
}

func SocketDisconnect(s socketio.Conn, reason string) {
	log.Trace("closed Socket.IO connection: ", reason)
}
