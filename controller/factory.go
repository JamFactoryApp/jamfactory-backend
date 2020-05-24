package controller

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"jamfactory-backend/helpers"
	"jamfactory-backend/middelwares"
	"net/http"
	"strings"
)

type joinBody struct {
	Label string `json:"label"`
}

func RegisterFactoryRoutes(router *mux.Router) {
	getSessionMiddleware := middelwares.GetSessionFromRequest{Store: Store}
	router.Use(getSessionMiddleware.Handler)

	router.HandleFunc("/create", createParty)

	joinBodyParser := middelwares.BodyParser{Body: new(joinBody)}
	router.Handle("/join", joinBodyParser.Handler(http.HandlerFunc(joinParty))).Methods("PUT")

	router.HandleFunc("/leave", leaveParty)
}

func createParty(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value("Session").(*sessions.Session)

	if !(session.Values["User"] != nil && session.Values["Token"] != nil && session.Values["User"] == "Host") {
		http.Error(w, "User Error: Not logged in to spotify", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not logged in to spotify ", session.ID)
		return
	}

	tokenMap := session.Values["Token"].(map[string]interface{})
	token := oauth2.Token{
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

	label, err := Factory.GenerateNewParty(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't create party: %s", session.ID, err.Error())
	}

	session.Values["Label"] = label
	helpers.SaveSession(w, r, session)

	res := make(map[string]interface{})
	res["label"] = label
	helpers.RespondWithJSON(w, res)
}

func joinParty(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value("Session").(*sessions.Session)
	body := r.Context().Value("Body").(*joinBody)

	party := Factory.GetParty(strings.ToUpper(body.Label))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	session.Values["User"] = "Guest"
	session.Values["Label"] = strings.ToUpper(body.Label)
	helpers.SaveSession(w, r, session)

	res := make(map[string]interface{})
	res["label"] = strings.ToUpper(body.Label)
	helpers.RespondWithJSON(w, res)
}

func leaveParty(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value("Session").(*sessions.Session)

	if session.Values["User"] != nil && session.Values["Label"] != nil && session.Values["User"] == "Host" {
		party := Factory.GetParty(session.Values["Label"].(string))
		if party != nil {
			party.SetPartyState(false)
		}
	}

	session.Values["User"] = "New"
	session.Values["Label"] = nil
	helpers.SaveSession(w, r, session)

	res := make(map[string]interface{})
	res["Success"] = true
	helpers.RespondWithJSON(w, res)
}
