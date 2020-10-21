package controllers

import (
	"context"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Middleware interface {
	Handler(http.Handler) http.Handler
}

type JamSessionRequiredMiddleware struct{}

type SessionRequiredMiddleware struct{}

type UserTypeRequiredMiddleware struct {
	UserType string
}

func (*JamSessionRequiredMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := utils.SessionFromRequestContext(r)

		if session == nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Fatal("Could not get session from context")
			return
		}

		logger := log.WithField("Session", session.ID)

		if !(session.Values[utils.SessionLabelTypeKey] != nil) {
			http.Error(w, "User error: Not joined a jamSession", http.StatusUnauthorized)
			logger.Trace("Could not get jamSession: User not joined a jamSession")
			return
		}

		jamSession := GetJamSession(session.Values[utils.SessionLabelTypeKey].(string))

		if jamSession == nil {
			http.Error(w, "JamSession error: Could not find a jamSession with the submitted label", http.StatusNotFound)
			logger.WithField("Label", session.Values[utils.SessionLabelTypeKey].(string)).Trace("Could not get jamSession: JamSession not found")
			return
		}

		ctx := context.WithValue(r.Context(), utils.JamSessionContextKey, jamSession)
		rWithCtx := r.WithContext(ctx)

		next.ServeHTTP(w, rWithCtx)
	})
}

func (*SessionRequiredMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r, "user-session")

		if session.IsNew {
			if err := session.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Debug("Could not save session")
				return
			}
		}

		ctx := context.WithValue(r.Context(), utils.SessionContextKey, session)
		rWithCtx := r.WithContext(ctx)

		next.ServeHTTP(w, rWithCtx)
	})
}

func (m *UserTypeRequiredMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := utils.SessionFromRequestContext(r)

		if !(session.Values[utils.SessionUserTypeKey] == m.UserType) {
			http.Error(w, "User Error: Not the correct user type", http.StatusUnauthorized)
			log.WithFields(log.Fields{
				"Current": session.Values[utils.SessionUserTypeKey],
				"Wanted":  m.UserType,
				"Session": session.ID,
			}).Debug("User Error: Not the correct user type")
			return
		}

		next.ServeHTTP(w, r)
	})
}
