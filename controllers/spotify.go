package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/utils"
	"net/http"
	"strings"
)

type searchRequestBody struct {
	SearchText string `json:"text"`
}

func devices(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	result, err := party.Client.PlayerDevices()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Party", party.Label).Debug("Could not get devices for party: ", err.Error())
		return
	}

	utils.EncodeJSONBody(w, result)
}

func playlist(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	result, err := party.Client.CurrentUsersPlaylists()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("Party", party.Label).Debug("Could not get playlists for party: ", err.Error())
		return
	}

	utils.EncodeJSONBody(w, result)
}

func search(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	var body searchRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
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

	utils.EncodeJSONBody(w, result)
}
