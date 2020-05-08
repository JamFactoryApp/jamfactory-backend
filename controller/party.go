package controller

import (
	"github.com/gorilla/mux"
	"net/http"
)

func RegisterPartyRoutes(router *mux.Router) {
	router.HandleFunc("/create", createParty)
	router.HandleFunc("/info", getPartyInfo)
	router.HandleFunc("/join", joinParty).Methods("PUT")
	router.HandleFunc("/leave", leaveParty)
	router.HandleFunc("/name", setPartyName).Methods("PUT")
	router.HandleFunc("/playback", setPlayback).Methods("PUT")
	router.HandleFunc("/playlist", addPlaylist).Methods("PUT")
	router.HandleFunc("/queue", getQueue)
	router.HandleFunc("/settings", setSettings).Methods("PUT")
	router.HandleFunc("/state", getPartyState)
	router.HandleFunc("/vote", vote).Methods("PUT")
}

func createParty(w http.ResponseWriter, r *http.Request) {

}

func getPartyInfo(w http.ResponseWriter, r *http.Request) {

}

func joinParty(w http.ResponseWriter, r *http.Request) {

}

func leaveParty(w http.ResponseWriter, r *http.Request) {

}

func setPartyName(w http.ResponseWriter, r *http.Request) {

}

func setPlayback(w http.ResponseWriter, r *http.Request) {

}

func addPlaylist(w http.ResponseWriter, r *http.Request) {

}

func getQueue(w http.ResponseWriter, r *http.Request) {

}

func setSettings(w http.ResponseWriter, r *http.Request) {

}

func getPartyState(w http.ResponseWriter, r *http.Request) {

}

func vote(w http.ResponseWriter, r *http.Request) {

}
