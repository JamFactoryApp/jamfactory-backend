package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
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
	session, err := Store.Get(r, "user-session")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsNew {
		session.Values["visits"] = int32(0)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	session.Values["visits"] = session.Values["visits"].(int32) + 1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if session.IsNew {
		_, err := fmt.Fprintf(w, "This is the first time you visited this site! Welcome")
		if err != nil {
			log.Println("Error printing to http.ResponseWriter")
		}
	}

	_, err = fmt.Fprintf(w, "You visited this site %d times ", session.Values["visits"].(int32))
	if err != nil {
		log.Println("Error printing to http.ResponseWriter")
	}
}

func pause(w http.ResponseWriter, r *http.Request) {

}

func play(w http.ResponseWriter, r *http.Request) {

}

func playlist(w http.ResponseWriter, r *http.Request) {

}

func search(w http.ResponseWriter, r *http.Request) {

}
