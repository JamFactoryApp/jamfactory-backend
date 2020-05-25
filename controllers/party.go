package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
	"net/http"
	"strings"
)

type partyBody struct {
	Name     string     `json:"name"`
	DeviceID spotify.ID `json:"device"`
	IpVoting bool       `json:"ip"`
}
type getPartyResponseBody partyBody
type setPartyRequestBody partyBody
type setPartyResponseBody partyBody

type playbackBody struct {
	CurrentSong *spotify.FullTrack   `json:"currentSong"`
	Playback    *spotify.PlayerState `json:"playback"`
}
type getPlaybackResponseBody playbackBody
type setPlayBackRequestBody playbackBody
type setPlaybackResponseBody playbackBody

type labelBody struct {
	Label string `json:"label"`
}
type createPartyResponseBody labelBody
type joinRequestBody labelBody
type joinResponseBody labelBody

type leavePartyResponseBody struct {
	Success bool `json:"Success"`
}

type partyStateResponseBody struct {
	CurrentSong interface{} `json:"currentSong"`
	State       interface{} `json:"state"`
}

func getParty(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	res := getPartyResponseBody{
		Name:     party.User.DisplayName,
		DeviceID: party.DeviceID,
		IpVoting: party.IpVoteEnabled,
	}

	utils.EncodeJSONBody(w, res)
}

func setParty(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	var body setPartyRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	party.SetClientID(body.DeviceID)
	party.IpVoteEnabled = body.IpVoting
	party.User.DisplayName = body.Name

	res := setPartyResponseBody{
		Name:     party.User.DisplayName,
		DeviceID: party.DeviceID,
		IpVoting: party.IpVoteEnabled,
	}

	utils.EncodeJSONBody(w, res)
}

func getPlayback(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	res := getPlaybackResponseBody{
		CurrentSong: party.CurrentSong,
		Playback:    party.PlaybackState,
	}

	utils.EncodeJSONBody(w, res)
}

func setPlayback(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	var body setPlayBackRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	party.SetPartyState(body.Playback.Playing)

	res := setPlaybackResponseBody{
		CurrentSong: party.CurrentSong,
		Playback:    party.PlaybackState,
	}

	utils.EncodeJSONBody(w, res)
}

func createParty(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if !LoggedInAsHost(session) {
		http.Error(w, "User Error: Not logged in to spotify", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not logged in to spotify ", session.ID)
		return
	}

	tokenMap := session.Values[SessionTokenKey].(map[string]interface{})
	token := oauth2.Token{
		AccessToken:  tokenMap["accesstoken"].(string),
		TokenType:    tokenMap["tokentype"].(string),
		RefreshToken: tokenMap["refreshtoken"].(string),
		Expiry:       tokenMap["expiry"].(primitive.DateTime).Time(),
	}

	if !token.Valid() {
		http.Error(w, "User Error: Token not valid", http.StatusUnauthorized)
		log.Printf("@%s User Error: Token not valid", session.ID)
		return
	}

	client := spotifyAuthenticator.NewClient(&token)

	label, err := GenerateNewParty(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't create party: %s", session.ID, err.Error())
	}

	session.Values[SessionLabelKey] = label
	SaveSession(w, r, session)

	res := createPartyResponseBody{Label: label}
	utils.EncodeJSONBody(w, res)
}

func joinParty(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	var body joinRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	party := GetParty(strings.ToUpper(body.Label))

	if party == nil {
		http.Error(w, "Party Error: Could not find a party with the submitted label", http.StatusNotFound)
		log.Printf("@%s Party Error: Could not find a party with the submitted label", session.ID)
		return
	}

	session.Values[SessionUserTypeKey] = models.UserTypeGuest
	session.Values[SessionLabelKey] = party.Label
	SaveSession(w, r, session)

	res := joinResponseBody{Label: party.Label}
	utils.EncodeJSONBody(w, res)
}

func leaveParty(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if LoggedInAsHost(session) {
		party := GetParty(session.Values[SessionLabelKey].(string))
		if party != nil {
			party.SetPartyState(false)
			body := partyStateResponseBody{
				CurrentSong: party.CurrentSong,
				State:       party.PlaybackState,
			}
			Socket.BroadcastToRoom("sessions", party.Label, SocketEventPlayback, body)
		}
	}

	session.Values[SessionUserTypeKey] = models.UserTypeNew
	session.Values[SessionLabelKey] = nil
	SaveSession(w, r, session)

	res := leavePartyResponseBody{Success: true}
	utils.EncodeJSONBody(w, res)
}
