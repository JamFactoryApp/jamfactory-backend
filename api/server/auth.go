package server

import (
	users2 "github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	"net/http"

	"github.com/pkg/errors"

	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
)

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	state := session.ID
	url := s.authenticator.CallbackURL(state)

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

	token, id, username, err := s.authenticator.Authenticate(session.ID, r)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	if state := r.FormValue("state"); state != session.ID {
		s.errNotFound(w, apierrors.ErrTokenMismatch, log.DebugLevel)
		return
	}

	user, err := s.users.Get(id)
	if err != nil {
		if errors.Is(err, users2.ErrUserNotFound) {
			user = users2.New(id, username, users2.UserTypeSpotify, token, s.authenticator)
		} else {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	} else {
		user.UserType = users2.UserTypeSpotify
		user.SpotifyToken = token
	}

	if err := s.users.Save(user, user.Identifier); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	sessions.SetIdentifier(session, id)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	origin, err := sessions.Origin(session)

	http.Redirect(w, r, origin, http.StatusSeeOther)
}
