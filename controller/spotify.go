package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"jamfactory-backend/models"
	"net/http"
)

type SpotifyEnv struct {
	*models.Env
}


func RegisterSpotifyRoutes(router *mux.Router, mainEnv *models.Env) {
	env := SpotifyEnv{mainEnv}
	router.HandleFunc("/devices", env.devices)
	router.HandleFunc("/pause", env.pause)
	router.HandleFunc("/play", env.play).Methods("PUT")
	router.HandleFunc("/playlist", env.playlist)
	router.HandleFunc("/search", env.search).Methods("PUT")
}

func (env *SpotifyEnv) devices(w http.ResponseWriter, r *http.Request) {

	session, err := env.Store.Get(r, "user-session")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew {
		session.Values["visits"] = 0
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	session.Values["visits"] = session.Values["visits"].(int) + 1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "You visited this site %d times", session.Values["visits"].(int))
}

func (env *SpotifyEnv) pause(w http.ResponseWriter, r *http.Request) {

}

func (env *SpotifyEnv) play(w http.ResponseWriter, r *http.Request) {

}

func (env *SpotifyEnv) playlist(w http.ResponseWriter, r *http.Request) {

}

func (env *SpotifyEnv) search(w http.ResponseWriter, r *http.Request) {

}
