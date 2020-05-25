package controllers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/utils"
	"net/http"
)

type Middleware interface {
	Handler(http.Handler) http.Handler
}

type PartyRequiredMiddleware struct{}

type SessionRequiredMiddleware struct{}

type UserTypeRequiredMiddleware struct {
	UserType string
}

func (*PartyRequiredMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := utils.SessionFromRequestContext(r)

		if session == nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Panic("Could not get session from context")
			return
		}

		logger := log.WithField("Session", session.ID)

		if !(session.Values[SessionLabelKey] != nil) {
			http.Error(w, "User error: Not joined a party", http.StatusUnauthorized)
			logger.Trace("Could not get party: User not joined a party")
			return
		}

		party := GetParty(session.Values[SessionLabelKey].(string))

		if party == nil {
			http.Error(w, "Party error: Could not find a party with the submitted label", http.StatusNotFound)
			logger.WithField("Label", session.Values[SessionLabelKey].(string)).Trace("Could not get party: Party not found")
			return
		}

		ctx := context.WithValue(r.Context(), utils.PartyContextKey, party)
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

		if !(session.Values[SessionUserTypeKey] == m.UserType) {
			http.Error(w, "User Error: Not the correct user type", http.StatusUnauthorized)
			log.WithFields(log.Fields{
				"Current": session.Values[SessionUserTypeKey],
				"Wanted":  m.UserType,
				"Session": session.ID,
			}).Debug("User Error: Not the correct user type")
			return
		}

		next.ServeHTTP(w, r)
	})
}
