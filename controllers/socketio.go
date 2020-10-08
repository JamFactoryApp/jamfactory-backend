package controllers

import (
	"errors"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	SocketNamespace = "/"

	SocketEventPlayback = "playback"
	SocketEventQueue    = "queue"
	SocketEventClose    = "close"

	CloseTypeHostLeft      = "host"
	CloseTypePartyInactive = "inactive"
)

func initSocketIO() {
	var err error

	pt := polling.Default
	wt := websocket.Default
	wt.CheckOrigin = func(req *http.Request) bool {
		return true
	}

	Socket, err = socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			pt,
			wt,
		},
	})
	if err != nil {
		log.Fatalln("Error creating socketio server\n", err)
	}

	go func() {
		defer utils.CloseProperly(Socket)
		err := Socket.Serve()
		if err != nil {
			log.Fatalf("Error serving socket.io server: %s\n", err)
		}
	}()
}

func socketIOConnect(s socketio.Conn) error {
	session, err := store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")

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

	if (session.Values[utils.SessionUserTypeKey] == models.UserTypeHost ||
		session.Values[utils.SessionUserTypeKey] == models.UserTypeGuest) &&
		session.Values[utils.SessionLabelTypeKey] != nil {

		if GetJamSession(session.Values[utils.SessionLabelTypeKey].(string)) != nil {
			s.Join(session.Values[utils.SessionLabelTypeKey].(string))
			s.SetContext("")
			logger.Trace("allowed connection")
			return nil
		}

		_ = s.Close()
		logger.Trace("disallowed connection: label invalid")
		return errors.New("label invalid")
	}

	_ = s.Close()
	logger.Trace("disallowed connection: wrong user type")
	return errors.New("wrong user type")
}

func socketIOError(s socketio.Conn, err error) {
	log.Error("Socket.IO Error:", err.Error())
}

func socketIODisconnect(s socketio.Conn, reason string) {
	log.Tracef("Closed Socket.IO connection %s: %s\n", s.ID(), reason)
}

func SendToRoom(room string, event string, message interface{}) {
	Socket.BroadcastToRoom(SocketNamespace, room, event, message)
}
