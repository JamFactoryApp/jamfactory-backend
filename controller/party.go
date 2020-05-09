package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"log"
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