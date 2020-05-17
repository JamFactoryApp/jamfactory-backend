package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"net/http"
	"strings"
)

func RegisterSpotifyRoutes(router *mux.Router) {
	router.HandleFunc("/devices", devices)
	router.HandleFunc("/playlist", playlist)
	router.HandleFunc("/search", search).Methods("PUT")
}

func devices(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	if !(session.Values["Label"] != nil) {
		http.Error(w, "User Error: Not joined a party", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not joined a party", session.ID)
		return
	}

	party := PartyControl.GetParty(session.Values["Label"].(string))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	result, err := party.Client.PlayerDevices()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't get devices: %s", session.ID, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func playlist(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	if !(session.Values["Label"] != nil) {
		http.Error(w, "User Error: Not joined a party", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not joined a party", session.ID)
		return
	}

	party := PartyControl.GetParty(session.Values["Label"].(string))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	result, err := party.Client.CurrentUsersPlaylists()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't get devices: %s", session.ID, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	decoder := json.NewDecoder(r.Body)

	var body struct{
		Text string `json:"text"`
	}

	err = decoder.Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't decode json from body: %s", session.ID, err.Error())
		return
	}

	if !(session.Values["Label"] != nil) {
		http.Error(w, "User Error: Not joined a party", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not joined a party", session.ID)
		return
	}

	party := PartyControl.GetParty(session.Values["Label"].(string))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	country := spotify.CountryGermany
	opts := spotify.Options{
		Country: &country,
	}
	searchString := []string{body.Text, "*"}
	result, err := party.Client.SearchOpt(strings.Join(searchString, ""), spotify.SearchTypeTrack, &opts)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't get search result: %s", session.ID, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(result)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}
