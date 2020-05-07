package controller

import (
	"errors"
	socketio "github.com/googollee/go-socket.io"
	"jamfactory-backend/models"
	"net/http"
)

type SocketEnv struct {
	*models.Env
}

func RegisterSocketRoutes(socket *socketio.Server, mainEnv *models.Env) {
	env := SocketEnv{mainEnv}

	socket.OnConnect("/", env.SocketAuth)
	socket.OnEvent("/", "connection", env.SocketConnect)

}

func (env *SocketEnv) SocketConnect(s socketio.Conn, msg string) {
	session, err := env.Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")

	s.Join(session.Values["label"].(string))
}

func (env *SocketEnv) SocketAuth(s socketio.Conn) error {
	session, err := env.Store.Get(&http.Request{Header: s.RemoteHeader()}, "user-session")
	if session.Values["user"] == "Host" || session.Values["user"] == "Guest" {
		if env.PartyController.GetParty(session.Values["label"].(string)) != nil {
			return nil
		} else {
			return errors.New("label is invalid")
		}
	}
	return errors.New("wrong user type")
}
