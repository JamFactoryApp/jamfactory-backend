package server

import (
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) current(w http.ResponseWriter, r *http.Request) {
	sessionType := s.CurrentSessionType(r)
	jamLabel := s.CurrentJamLabel(r)
	identifier := s.CurrentIdentifier(r)

	authorized := false
	if user, err := s.users.Get(identifier); err == nil {
		switch user.UserType {
		case types.UserTypeSpotify:
			if user.Token != nil && user.Token.Valid() {
				authorized = true
			}
		}
	}

	utils.EncodeJSONBody(w, types.GetAuthCurrentResponse{
		UserType:   string(sessionType),
		Label:      jamLabel,
		Authorized: authorized,
	})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	state := session.ID
	url := s.jamFactory.CallbackURL(state)

	sessions.SetOrigin(session, r.Header.Get("Referer"))

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	utils.EncodeJSONBody(w, types.GetAuthLoginResponse{
		URL: url,
	})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	identifier := s.CurrentIdentifier(r)
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	s.users.Delete(identifier)

	utils.EncodeJSONBody(w, types.GetAuthLogoutResponse{
		Success: true,
	})
}

func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)

	token, id, username, err := s.jamFactory.Authenticate(session.ID, r)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	if state := r.FormValue("state"); state != session.ID {
		s.errNotFound(w, apierrors.ErrTokenMismatch, log.DebugLevel)
		return
	}

	user := s.users.New(id, username, types.UserTypeSpotify, token)
	if err := s.users.Save(user); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	sessions.SetIdentifier(session, id)
	sessions.SetSessionType(session, types.SessionTypeNew)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	origin, err := sessions.Origin(session)

	http.Redirect(w, r, origin, http.StatusSeeOther)
}
