package server

import (
	"github.com/jamfactoryapp/jamfactory-backend/pkg/hub"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
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

	s.users.DeleteUser(identifier)

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

	user, err := s.users.GetUserByIdentifier(id)
	if err != nil {
		if errors.Is(err, hub.ErrUserNotFound) {
			user, err = s.users.NewUser(id, username, users.UserTypeSpotify, token)
			if err != nil {
				s.errInternalServerError(w, err, log.DebugLevel)
				return
			}
		} else {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	} else {
		userInfo, err := user.GetInfo()
		if err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
		userInfo.UserType = users.UserTypeSpotify
		userInfo.SpotifyToken = token
		if err := user.SetInfo(userInfo); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}

	}

	sessions.SetIdentifier(session, id)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	origin, err := sessions.Origin(session)

	http.Redirect(w, r, origin, http.StatusSeeOther)
}
