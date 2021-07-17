package server

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) websocketHandler(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.errInternalServerError(w, err, log.ErrorLevel)
		return
	}

	jamSession.IntroduceClient(conn)
}
