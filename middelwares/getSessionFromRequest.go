package middelwares

import (
	"context"
	log "github.com/sirupsen/logrus"
	"jamfactory-backend/models"
	"net/http"
)

type GetSessionFromRequest struct {
	Store *models.SessionStore
}

func (middleware *GetSessionFromRequest) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := middleware.Store.Get(r, "user-session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Debug("Could not get session for request")
			return
		}

		ctx := context.WithValue(r.Context(), "Session", session)
		rWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, rWithCtx)
	})
}
