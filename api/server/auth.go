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
	userType := s.CurrentUserType(r)
	jamLabel := s.CurrentJamLabel(r)
	token := s.CurrentToken(r)

	utils.EncodeJSONBody(w, types.GetAuthCurrentResponse{
		User:       string(userType),
		Label:      jamLabel,
		Authorized: token != nil && token.Valid(),
	})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	state := session.ID
	url := s.jamFactory.CallbackURL(state)

	utils.EncodeJSONBody(w, types.GetAuthLoginResponse{
		URL: url,
	})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	session.Options.MaxAge = -1

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	utils.EncodeJSONBody(w, types.GetAuthLogoutResponse{
		Success: true,
	})
}

func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)

	token, err := s.jamFactory.Authenticate(session.ID, r)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	if st := r.FormValue("state"); st != session.ID {
		s.errNotFound(w, apierrors.ErrTokenMismatch, log.DebugLevel)
		return
	}

	sessions.SetToken(session, token)
	sessions.SetUserType(session, types.UserTypeNew)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	http.Redirect(w, r, utils.JamRedirectURL(), http.StatusSeeOther)
}
