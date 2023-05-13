package server

import (
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/permissions"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	sessionCookieKey = "user-session"
)

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.config.Development {
			w.Header().Add("Access-Control-Allow-Origin", s.config.ClientAddresses[0].String())
			w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
			w.Header().Add("Access-Control-Allow-Credentials", "true")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" {
			for i := range s.config.ClientAddresses {
				if s.config.ClientAddresses[i].String() == origin {
					w.Header().Add("Access-Control-Allow-Origin", s.config.ClientAddresses[i].String())
					w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
					w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, DELETE, OPTIONS")
					w.Header().Add("Access-Control-Allow-Credentials", "true")
					if r.Method == "OPTIONS" {
						w.WriteHeader(http.StatusOK)
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}
		}
		log.Trace("Cors-Middleware Not-Allowed ")
		w.WriteHeader(http.StatusUnauthorized)
		return
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
		members, err := jamSession.GetMembers()
		if err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
		member, err := members.Get(user.Identifier)
		if err != nil || !member.HasPermissions(permissions.Host) {
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
