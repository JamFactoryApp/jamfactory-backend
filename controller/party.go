package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/zmb3/spotify"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"jamfactory-backend/models"
	"log"
	"net/http"
	"strings"
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
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !(session.Values["User"] != nil && session.Values["Token"] != nil && session.Values["User"] == "Host") {
		http.Error(w, "User Error: Not logged in to spotify", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not logged in to spotify ", session.ID)
		return
	}

	tokenMap := session.Values["Token"].(map[string]interface{})
	token  := oauth2.Token{
		AccessToken:  tokenMap["accesstoken"].(string),
		TokenType:    tokenMap["tokentype"].(string),
		RefreshToken: tokenMap["refreshtoken"].(string),
		Expiry:       tokenMap["expiry"].(primitive.DateTime).Time(),
	}

	if !(token.Valid() == true) {
		http.Error(w, "User Error: Token not valid", http.StatusUnauthorized)
		log.Printf("@%s User Error: Token not valid", session.ID)
		return
	}

	client := SpotifyAuthenticator.NewClient(&token)
	user, err := client.CurrentUser()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Could not create get current user: %s", session.ID, err.Error())
		return
	}

	label := PartyControl.generateNewParty(client, user)

	session.Values["Label"] = label

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Could not save session: %s", session.ID, err.Error())
		return
	}

	res := make(map[string]interface{})
	res["label"] = label
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func getPartyInfo(w http.ResponseWriter, r *http.Request) {
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

	res := make(map[string]interface{})
	res["id"] = party.User.ID
	res["display_name"] = party.User.DisplayName

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}

}

func joinParty(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	decoder := json.NewDecoder(r.Body)

	var body struct{
		Label string `json:"label"`
	}

	err = decoder.Decode(&body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't decode json from body: %s", session.ID, err.Error())
		return
	}

	party := PartyControl.GetParty(strings.ToUpper(body.Label))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	session.Values["User"] = "Guest"
	session.Values["Label"] = strings.ToUpper(body.Label)

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Could not save session: %s", session.ID, err.Error())
		return
	}

	res := make(map[string]interface{})
	res["label"] = strings.ToUpper(body.Label)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func leaveParty(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	if session.Values["User"] != nil && session.Values["Label"] != nil && session.Values["User"] == "Host" {
		party := PartyControl.GetParty(session.Values["Label"].(string))
		if party != nil {
			party.SetQueueActive(false)
		}
	}

	session.Values["User"] = "New"
	session.Values["Label"] = nil

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Could not save session: %s", session.ID, err.Error())
		return
	}

	res := make(map[string]interface{})
	res["Success"] = true

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}

}

func setPartyName(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	decoder := json.NewDecoder(r.Body)

	var body struct{
		Name string `json:"name"`
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

	party.User.DisplayName = body.Name

	res := make(map[string]interface{})
	res["id"] = party.User.ID
	res["display_name"] = party.User.DisplayName

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func setPlayback(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	decoder := json.NewDecoder(r.Body)

	var body struct{
		Playback bool `json:"playback"`
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

	party.SetQueueActive(body.Playback)

	res := make(map[string]interface{})
	res["Settings"] = "Saved"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func addPlaylist(w http.ResponseWriter, r *http.Request) {

}

func getQueue(w http.ResponseWriter, r *http.Request) {

}

func setSettings(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "user-session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println("Couldn't get session")
		return
	}

	decoder := json.NewDecoder(r.Body)

	var body struct{
		DeviceID spotify.ID `json:"device"`
		IpVoting bool `json:"ip"`
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

	settings := models.PartySettings{
		DeviceId: body.DeviceID,
		IpVoting: body.IpVoting,
	}

	party.SetSetting(settings)

	res := make(map[string]interface{})
	res["Settings"] = "Saved"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't encode json: %s", session.ID, err.Error())
	}
}

func getPartyState(w http.ResponseWriter, r *http.Request) {

}

func vote(w http.ResponseWriter, r *http.Request) {

}
