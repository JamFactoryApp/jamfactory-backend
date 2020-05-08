package controller

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterSpotifyRoutes(router *mux.Router) {
	router.HandleFunc("/devices", devices)
	router.HandleFunc("/pause", pause)
	router.HandleFunc("/play", play).Methods("PUT")
	router.HandleFunc("/playlist", playlist)
	router.HandleFunc("/search", search).Methods("PUT")
}

func devices(w http.ResponseWriter, r *http.Request) {

}

func pause(w http.ResponseWriter, r *http.Request) {

}

func play(w http.ResponseWriter, r *http.Request) {

}

func playlist(w http.ResponseWriter, r *http.Request) {

}

func search(w http.ResponseWriter, r *http.Request) {

}
