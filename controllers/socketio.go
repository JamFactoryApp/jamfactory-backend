package controllers

import (
	"errors"
	socketio "github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
	"net/http"
)

const (
	SocketEventPlayback = "playback"
	SocketEventQueue    = "queue"
)

func initSocketIO() {
	var err error
	Socket, err = socketio.NewServer(nil)
	if err != nil {
		log.Fatalln("Error creating socketio server\n", err)
	}

	go func() {
		defer utils.CloseProperly(Socket)
		err := Socket.Serve()
		if err != nil {
			log.Panicf("Error serving socket.io server:\n%s\n", err)
		}
	}()
}

func socketIOConnect(s socketio.Conn) error {
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

	if (session.Values[SessionUserTypeKey] == models.UserTypeHost ||
		session.Values[SessionUserTypeKey] == models.UserTypeGuest) &&
		session.Values[SessionLabelKey] != nil {

		if GetParty(session.Values[SessionLabelKey].(string)) != nil {
			s.Join(session.Values[SessionLabelKey].(string))
			s.SetContext(s.Context())
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
	log.Errorf("Socket.IO Error in %s:\n%s\n", s.ID(), err.Error())
}

func socketIODisconnect(s socketio.Conn, reason string) {
	log.Tracef("Closed Socket.IO connection %s:\n%s\n", s.ID(), reason)
}
