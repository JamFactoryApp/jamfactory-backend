package helpers

import (
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func SaveSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) {

	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Session", session.ID).Warn("Could not save session: ", err.Error())
		return
	}
}
