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

type jamSessionBody struct {
	Name     string     `json:"name"`
	DeviceID spotify.ID `json:"device"`
	IpVoting bool       `json:"ip"`
}
type getJamSessionResponseBody jamSessionBody
type setJamSessionRequestBody jamSessionBody
type setJamSessionResponseBody jamSessionBody

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
type createJamSessionResponseBody labelBody
type joinRequestBody labelBody
type joinResponseBody labelBody

type leaveJamSessionResponseBody struct {
	Success bool `json:"success"`
}

type jamSessionStateResponseBody struct {
	CurrentSong interface{} `json:"currentSong"`
	State       interface{} `json:"state"`
}

func getJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	res := getJamSessionResponseBody{
		Name:     jamSession.User.DisplayName,
		DeviceID: jamSession.DeviceID,
		IpVoting: jamSession.IpVoteEnabled,
	}

	utils.EncodeJSONBody(w, res)
}

func setJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body setJamSessionRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession.SetClientID(body.DeviceID)
	jamSession.IpVoteEnabled = body.IpVoting
	jamSession.User.DisplayName = body.Name

	res := setJamSessionResponseBody{
		Name:     jamSession.User.DisplayName,
		DeviceID: jamSession.DeviceID,
		IpVoting: jamSession.IpVoteEnabled,
	}

	utils.EncodeJSONBody(w, res)
}

func getPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	res := getPlaybackResponseBody{
		CurrentSong: jamSession.CurrentSong,
		Playback:    jamSession.PlaybackState,
	}

	utils.EncodeJSONBody(w, res)
}

func setPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body setPlayBackRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession.SetJamSessionState(body.Playback.Playing)

	res := setPlaybackResponseBody{
		CurrentSong: jamSession.CurrentSong,
		Playback:    jamSession.PlaybackState,
	}

	utils.EncodeJSONBody(w, res)
}

func createJamSession(w http.ResponseWriter, r *http.Request) {
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

	label, err := GenerateNewJamSession(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't create jamSession: %s", session.ID, err.Error())
	}

	session.Values[SessionLabelKey] = label
	SaveSession(w, r, session)

	res := createJamSessionResponseBody{Label: label}
	utils.EncodeJSONBody(w, res)
}

func joinJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	var body joinRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := GetJamSession(strings.ToUpper(body.Label))

	if jamSession == nil {
		http.Error(w, "JamSession Error: Could not find a jamSession with the submitted label", http.StatusNotFound)
		log.Printf("@%s JamSession Error: Could not find a jamSession with the submitted label", session.ID)
		return
	}

	session.Values[SessionUserTypeKey] = models.UserTypeGuest
	session.Values[SessionLabelKey] = jamSession.Label
	SaveSession(w, r, session)

	res := joinResponseBody{Label: jamSession.Label}
	utils.EncodeJSONBody(w, res)
}

func leaveJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if LoggedInAsHost(session) {
		jamSession := GetJamSession(session.Values[SessionLabelKey].(string))
		if jamSession != nil {
			jamSession.SetJamSessionState(false)
			body := jamSessionStateResponseBody{
				CurrentSong: jamSession.CurrentSong,
				State:       jamSession.PlaybackState,
			}
			Socket.BroadcastToRoom("sessions", jamSession.Label, SocketEventPlayback, body)
		}
	}

	session.Values[SessionUserTypeKey] = models.UserTypeNew
	session.Values[SessionLabelKey] = nil
	SaveSession(w, r, session)

	res := leaveJamSessionResponseBody{Success: true}
	utils.EncodeJSONBody(w, res)
}
