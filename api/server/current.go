package server

import (
	"github.com/gorilla/sessions"
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	pkgsessions "github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"golang.org/x/oauth2"
	"net/http"
)

func (s *Server) CurrentSession(r *http.Request) *sessions.Session {
	session, err := pkgsessions.FromContext(r.Context())
	if err != nil {
		panic(err)
	}
	return session
}

func (s *Server) CurrentJamLabel(r *http.Request) string {
	session := s.CurrentSession(r)
	jamLabel, err := pkgsessions.JamLabel(session)
	if err != nil {
		return ""
	}
	return jamLabel
}

func (s *Server) CurrentJamSession(r *http.Request) jamsession.JamSession {
	jamSession, err := jamsession.FromContext(r.Context())
	if err != nil {
		panic(err)
	}
	return jamSession
}

func (s *Server) CurrentToken(r *http.Request) *oauth2.Token {
	session := s.CurrentSession(r)
	token, err := pkgsessions.Token(session)
	if err != nil {
		return nil
	}
	return token
}

func (s *Server) CurrentUserType(r *http.Request) types.UserType {
	session := s.CurrentSession(r)
	userType, err := pkgsessions.UserType(session)
	if err != nil {
		return types.UserTypeNew
	}
	return userType
}

func (s *Server) CurrentVoteID(r *http.Request) string {
	jamSession := s.CurrentJamSession(r)

	var voteID string

	switch jamSession.VotingType() {
	case types.SessionVoting:
		session := s.CurrentSession(r)
		voteID = session.ID
	case types.IPVoting:
		voteID = r.RemoteAddr
	default:
		panic(apierrors.ErrInvalidVotingType)
	}

	return voteID
}
