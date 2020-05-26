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
	jamSession := utils.JamSessionFromRequestContext(r)

	result, err := jamSession.Client.PlayerDevices()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("JamSession", jamSession.Label).Debug("Could not get devices for jamSession: ", err.Error())
		return
	}

	utils.EncodeJSONBody(w, result)
}

func playlist(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	result, err := jamSession.Client.CurrentUsersPlaylists()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithField("JamSession", jamSession.Label).Debug("Could not get playlists for jamSession: ", err.Error())
		return
	}

	utils.EncodeJSONBody(w, result)
}

func search(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body searchRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	country := spotify.CountryGermany
	opts := spotify.Options{
		Country: &country,
	}

	searchString := []string{body.SearchText, "*"}
	result, err := jamSession.Client.SearchOpt(strings.Join(searchString, ""), spotify.SearchTypeTrack, &opts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"JamSession": jamSession.Label,
			"Text":       body.SearchText}).Debug("Could not get search results: ", err.Error())
		return
	}

	utils.EncodeJSONBody(w, result)
}
