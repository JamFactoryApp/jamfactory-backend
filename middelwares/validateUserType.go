package middelwares

import (
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"net/http"
)

type ValidateUserType struct {
	User string
}

func (middleware *ValidateUserType) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session := r.Context().Value("Session").(*sessions.Session)

		if !(session.Values[models.SessionUserKey] == middleware.User) {
			http.Error(w, "User Error: Not the correct user", http.StatusUnauthorized)
			log.WithFields(log.Fields{
				"Current": session.Values[models.SessionUserKey],
				"Wanted":  middleware.User,
				"Session": session.ID,
			}).Debug("User Error: Not the correct user")
			return
		}
		next.ServeHTTP(w, r)
	})
}
