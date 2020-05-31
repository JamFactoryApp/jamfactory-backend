package controllers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
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
			log.Panic("Could not get session from context")
			return
		}

		logger := log.WithField("Session", session.ID)

		if !(session.Values[models.SessionLabelTypeKey] != nil) {
			http.Error(w, "User error: Not joined a jamSession", http.StatusUnauthorized)
			logger.Trace("Could not get jamSession: User not joined a jamSession")
			return
		}

		jamSession := GetJamSession(session.Values[models.SessionLabelTypeKey].(string))

		if jamSession == nil {
			http.Error(w, "JamSession error: Could not find a jamSession with the submitted label", http.StatusNotFound)
			logger.WithField("Label", session.Values[models.SessionLabelTypeKey].(string)).Trace("Could not get jamSession: JamSession not found")
			return
		}

		ctx := context.WithValue(r.Context(), utils.JamSessionContextKey, jamSession)
		rWithCtx := r.WithContext(ctx)

		next.ServeHTTP(w, rWithCtx)
	})
}

func (*SessionRequiredMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSession(r, "user-session")

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Debug("Could not get session for request")
			return
		}

		ctx := context.WithValue(r.Context(), utils.SessionContextKey, session)
		rWithCtx := r.WithContext(ctx)

		next.ServeHTTP(w, rWithCtx)
	})
}

func (m *UserTypeRequiredMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := utils.SessionFromRequestContext(r)

		if !(session.Values[models.SessionUserTypeKey] == m.UserType) {
			http.Error(w, "User Error: Not the correct user type", http.StatusUnauthorized)
			log.WithFields(log.Fields{
				"Current": session.Values[models.SessionUserTypeKey],
				"Wanted":  m.UserType,
				"Session": session.ID,
			}).Debug("User Error: Not the correct user type")
			return
		}

		next.ServeHTTP(w, r)
	})
}
