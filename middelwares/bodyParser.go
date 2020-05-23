package middelwares

import (
	"context"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type BodyParser struct {
	Body interface{}
}

func (middleware *BodyParser) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&middleware.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			log.Warn("Could not decode json from body: ", err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), "Body", middleware.Body)
		rWithCtx := r.WithContext(ctx)
		next.ServeHTTP(w, rWithCtx)
	})
}
