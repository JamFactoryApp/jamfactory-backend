package controller

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	chain "github.com/justinas/alice"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"jamfactory-backend/helpers"
	"jamfactory-backend/middelwares"
	"jamfactory-backend/models"
	"net/http"
	"strings"
)

type joinBody struct {
	Label string `json:"label"`
}

func RegisterFactoryRoutes(router *mux.Router) {
	getSessionMiddleware := middelwares.GetSessionFromRequest{Store: Store}
	stdChain := chain.New(getSessionMiddleware.Handler)

	router.Handle("/create", stdChain.ThenFunc(createParty)).Methods("GET")
	router.Handle("/join", stdChain.ThenFunc(joinParty)).Methods("PUT")
	router.Handle("/leave", stdChain.ThenFunc(leaveParty)).Methods("GET")
}

func createParty(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(models.SessionContextKey).(*sessions.Session)

	if !(session.Values[models.SessionUserKey] != nil && session.Values[models.SessionTokenKey] != nil && session.Values[models.SessionUserKey] == models.UserTypeHost) {
		http.Error(w, "User Error: Not logged in to spotify", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not logged in to spotify ", session.ID)
		return
	}

	tokenMap := session.Values[models.SessionTokenKey].(map[string]interface{})
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

	session.Values[models.SessionLabelKey] = label
	helpers.SaveSession(w, r, session)

	res := make(map[string]interface{})
	res["label"] = label
	helpers.RespondWithJSON(w, res)
}

func joinParty(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(models.SessionContextKey).(*sessions.Session)

	var body joinBody
	if err := helpers.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	party := Factory.GetParty(strings.ToUpper(body.Label))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	session.Values[models.SessionUserKey] = models.UserTypeGuest
	session.Values[models.SessionLabelKey] = strings.ToUpper(body.Label)
	helpers.SaveSession(w, r, session)

	res := make(map[string]interface{})
	res["label"] = strings.ToUpper(body.Label)
	helpers.RespondWithJSON(w, res)
}

func leaveParty(w http.ResponseWriter, r *http.Request) {
	session := r.Context().Value(models.SessionContextKey).(*sessions.Session)

	if session.Values[models.SessionUserKey] != nil && session.Values[models.SessionLabelKey] != nil && session.Values[models.SessionUserKey] == models.UserTypeHost {
		party := Factory.GetParty(session.Values[models.SessionLabelKey].(string))
		if party != nil {
			party.SetPartyState(false)
		}
	}

	session.Values[models.SessionUserKey] = models.UserTypeNew
	session.Values[models.SessionLabelKey] = nil
	helpers.SaveSession(w, r, session)

	res := make(map[string]interface{})
	res["Success"] = true
	helpers.RespondWithJSON(w, res)
}
