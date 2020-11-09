package server

import (
	"github.com/jamfactoryapp/jamfactory-backend/internal/logutils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) error(w http.ResponseWriter, err error, code int, level log.Level) {
	logutils.Log(level, err.Error())
	http.Error(w, err.Error(), code)
}

func (s *Server) errBadRequest(w http.ResponseWriter, err error, level log.Level) {
	s.error(w, err, http.StatusBadRequest, level)
}

func (s *Server) errUnauthorized(w http.ResponseWriter, err error, level log.Level) {
	s.error(w, err, http.StatusUnauthorized, level)
}

func (s *Server) errForbidden(w http.ResponseWriter, err error, level log.Level) {
	s.error(w, err, http.StatusForbidden, level)
}

func (s *Server) errNotFound(w http.ResponseWriter, err error, level log.Level) {
	s.error(w, err, http.StatusNotFound, level)
}

func (s *Server) errInternalServerError(w http.ResponseWriter, err error, level log.Level) {
	s.error(w, err, http.StatusInternalServerError, level)
}

func (s *Server) errSession(w http.ResponseWriter, err error) {
	s.error(w, err, http.StatusBadRequest, log.WarnLevel)
}

func (s *Server) errSessionSave(w http.ResponseWriter, err error) {
	s.error(w, err, http.StatusInternalServerError, log.ErrorLevel)
}
