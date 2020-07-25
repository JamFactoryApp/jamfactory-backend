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
	SearchType string `json:"type"`
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
	var searchType spotify.SearchType
	switch body.SearchType {
	case "track":
		searchType = spotify.SearchTypeTrack
	case "playlist":
		searchType = spotify.SearchTypePlaylist
	case "album":
		searchType = spotify.SearchTypeAlbum
	}

	if searchType == 0 {
		http.Error(w, "Unsupported search type", http.StatusUnprocessableEntity)
		log.WithFields(log.Fields{
			"JamSession": jamSession.Label,
			"Text":       body.SearchText}).Debug("Unsupported search type: ", body.SearchType)
		return
	}

	searchString := []string{body.SearchText, "*"}
	result, err := jamSession.Client.SearchOpt(strings.Join(searchString, ""), searchType, &opts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.WithFields(log.Fields{
			"JamSession": jamSession.Label,
			"Text":       body.SearchText}).Debug("Could not get search results: ", err.Error())
		return
	}

	utils.EncodeJSONBody(w, result)
}
