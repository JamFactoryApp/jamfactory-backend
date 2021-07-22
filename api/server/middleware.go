package server

import (
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/users"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	sessionCookieKey = "user-session"
)

func (s *Server) corsMiddleware(next http.Handler, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Add("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
			w.Header().Add("Access-Control-Allow-Credentials", "true")
		}
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) sessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.store.Get(r, sessionCookieKey)

		if err != nil {
			log.Debug(err)
		}

		if session.IsNew {
			if err := session.Save(r, w); err != nil {
				s.errInternalServerError(w, apierrors.ErrSessionCouldNotSave, log.ErrorLevel)
				return
			}
		}

		ctx := sessions.NewContext(r.Context(), session)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		identifier := s.CurrentIdentifier(r)
		user, err := s.users.Get(identifier)
		if err != nil {
			user = s.users.NewEmpty()
		}
		ctx := users.NewContext(r.Context(), user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) jamSessionRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.CurrentUser(r)

		jamSession, err := s.jamFactory.GetJamSessionByUser(user)
		if err != nil {
			s.errUnauthorized(w, err, log.TraceLevel)
			return
		}

		ctx := jamsession.NewContext(r.Context(), jamSession)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (s *Server) hostRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.CurrentUser(r)
		jamSession := s.CurrentJamSession(r)
		member, err := jamSession.Members().Get(user)
		if err != nil || !member.Has([]types.MemberRights{types.RightHost}) {
			s.errUnauthorized(w, apierrors.ErrUserTypeInvalid, log.DebugLevel)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) nonMemberRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.CurrentUser(r)
		if _, err := s.jamFactory.GetJamSessionByUser(user); err == nil {
			s.errUnauthorized(w, apierrors.ErrAlreadyMember, log.DebugLevel)
			return
		}
		next.ServeHTTP(w, r)
	})
}
