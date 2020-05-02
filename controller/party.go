package controller

import (
	"fmt"
	"github.com/gorilla/mux"
	"jamfactory-backend/models"
	"net/http"
)

type PartyEnv struct {
	*models.Env
}

func RegisterPartyRoutes(router *mux.Router, mainEnv *models.Env) {
	env := PartyEnv{mainEnv}
	router.HandleFunc("/create", env.createParty)
	router.HandleFunc("/info", env.getPartyInfo)
	router.HandleFunc("/join", env.joinParty).Methods("PUT")
	router.HandleFunc("/leave", env.leaveParty)
	router.HandleFunc("/name", env.setPartyName).Methods("PUT")
	router.HandleFunc("/playback", env.setPlayback).Methods("PUT")
	router.HandleFunc("/playlist", env.addPlaylist).Methods("PUT")
	router.HandleFunc("/queue", env.getQueue)
	router.HandleFunc("/settings", env.setSettings).Methods("PUT")
	router.HandleFunc("/state", env.getPartyState)
	router.HandleFunc("/vote", env.vote).Methods("PUT")
}

func (env *PartyEnv) createParty(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) getPartyInfo(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) joinParty(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "you are joining yay")
}

func (env *PartyEnv) leaveParty(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) setPartyName(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) setPlayback(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) addPlaylist(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) getQueue(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) setSettings(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) getPartyState(w http.ResponseWriter, r *http.Request) {

}

func (env *PartyEnv) vote(w http.ResponseWriter, r *http.Request) {

}
