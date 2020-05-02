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
	fmt.Fprintf(w, env.Text)
}

func (env *SpotifyEnv) pause(w http.ResponseWriter, r *http.Request) {

}

func (env *SpotifyEnv) play(w http.ResponseWriter, r *http.Request) {

}

func (env *SpotifyEnv) playlist(w http.ResponseWriter, r *http.Request) {

}

func (env *SpotifyEnv) search(w http.ResponseWriter, r *http.Request) {

}
