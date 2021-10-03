package server

import (
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)

	var jamLabel string
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err != nil {
		jamLabel = ""
	} else {
		jamLabel = jamSession.JamLabel()
	}

	spotifyAuthorized := false
	if user.SpotifyToken != nil && user.SpotifyToken.Valid() {
		spotifyAuthorized = true
	}

	utils.EncodeJSONBody(w, types.GetUserResponse{
		Identifier:        user.Identifier,
		DisplayName:       user.UserName,
		UserType:          string(user.UserType),
		JoinedLabel:       jamLabel,
		SpotifyAuthorized: spotifyAuthorized,
	})
}

func (s *Server) setUser(w http.ResponseWriter, r *http.Request) {
	var body types.PutUserRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}
	user := s.CurrentUser(r)
	user.UserName = body.DisplayName
	err := s.users.Save(user)
	if err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	var jamLabel string
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err != nil {
		jamLabel = ""
	} else {
		jamLabel = jamSession.JamLabel()
	}

	spotifyAuthorized := false
	if user.SpotifyToken != nil && user.SpotifyToken.Valid() {
		spotifyAuthorized = true
	}

	utils.EncodeJSONBody(w, types.GetUserResponse{
		Identifier:        user.Identifier,
		DisplayName:       user.UserName,
		UserType:          string(user.UserType),
		JoinedLabel:       jamLabel,
		SpotifyAuthorized: spotifyAuthorized,
	})
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	err := s.users.Delete(user.Identifier)
	if err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}
	utils.EncodeJSONBody(w, types.DeleteUserResponse{
		Success: true,
	})
}
