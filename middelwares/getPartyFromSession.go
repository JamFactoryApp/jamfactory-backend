package middelwares

import (
	"context"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"net/http"
)

type GetPartyFromSession struct {
	PartyControl *models.Factory
}

func (middleware *GetPartyFromSession) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session := r.Context().Value("Session").(*sessions.Session)
		if session == nil {
			http.Error(w, "", http.StatusInternalServerError)
			log.Panic("Could not get Session From Context")
			return
		}

		logger := log.WithField("Session", session.ID)

		if !(session.Values["Label"] != nil) {
			http.Error(w, "User Error: Not joined a party", http.StatusUnauthorized)
			logger.Trace("Could not get party: User not joined a party")
			return
		}

		party := middleware.PartyControl.GetParty(session.Values["Label"].(string))

		if party == nil {
			http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
			logger.WithField("Label", session.Values["Label"].(string)).Trace("Could not get party: Party not found")
			return
		}

		ctx := context.WithValue(r.Context(), "Party", party)
		rWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, rWithCtx)
	})
}
