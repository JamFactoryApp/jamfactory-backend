package server

import (
	"github.com/gorilla/sessions"
	pkgsessions "github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/users"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"net/http"
	"time"
)

func (s *Server) CurrentSession(r *http.Request) *sessions.Session {
	session, err := pkgsessions.FromContext(r.Context())
	if err != nil {
		panic(err)
	}
	return session
}

func (s *Server) CurrentUser(r *http.Request) *users.User {
	user, err := users.FromContext(r.Context())
	if err != nil {
		panic(err)
	}
	return user
}

func (s *Server) CurrentJamSession(r *http.Request) jamsession.JamSession {
	jamSession, err := jamsession.FromContext(r.Context())
	if err != nil {
		panic(err)
	}
	jamSession.SetTimestamp(time.Now())
	return jamSession
}

func (s *Server) CurrentIdentifier(r *http.Request) string {
	session := s.CurrentSession(r)
	id, err := pkgsessions.Identifier(session)
	if err != nil {
		return ""
	}
	return id
}

func (s *Server) CurrentVoteID(r *http.Request) string {
	user := s.CurrentUser(r)
	return user.Identifier
}
