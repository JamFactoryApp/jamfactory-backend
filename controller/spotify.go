package controller

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/helpers"
	"jamfactory-backend/middelwares"
	"jamfactory-backend/models"
	"net/http"
	"strings"
)

func RegisterSpotifyRoutes(router *mux.Router) {
	getSessionMiddleware := middelwares.GetSessionFromRequest{Store: Store}
	getPartyMiddleware := middelwares.GetPartyFromSession{PartyControl: &Factory}

	stdChain := alice.New(getSessionMiddleware.Handler, getPartyMiddleware.Handler)

	router.Handle("/devices", stdChain.ThenFunc(devices)).Methods("GET")
	router.Handle("/playlist", stdChain.ThenFunc(playlist)).Methods("GET")
	router.Handle("/search", stdChain.ThenFunc(search)).Methods("PUT")
}

type searchBody struct {
	SearchText string `json:"text"`
}

func devices(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)

	result, err := party.Client.PlayerDevices()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Party", party.Label).Debug("Could not get devices for party: ", err.Error())
		return
	}

	helpers.RespondWithJSON(w, result)
}

func playlist(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)

	result, err := party.Client.CurrentUsersPlaylists()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Party", party.Label).Debug("Could not get playlists for party: ", err.Error())
		return
	}

	helpers.RespondWithJSON(w, result)
}

func search(w http.ResponseWriter, r *http.Request) {
	party := r.Context().Value("Party").(*models.Party)

	var body searchBody
	if err := helpers.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	country := spotify.CountryGermany
	opts := spotify.Options{
		Country: &country,
	}

	searchString := []string{body.SearchText, "*"}
	result, err := party.Client.SearchOpt(strings.Join(searchString, ""), spotify.SearchTypeTrack, &opts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"Party": party.Label,
			"Text":  body.SearchText}).Debug("Could not get search results: ", err.Error())
		return
	}

	helpers.RespondWithJSON(w, result)
}
