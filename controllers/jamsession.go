package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
	"net/http"
	"strings"
)

type jamSessionBody struct {
	Label    string     `json:"label"`
	Name     string     `json:"name"`
	Active   bool       `json:"active"`
	DeviceID spotify.ID `json:"device_id"`
	IpVoting bool       `json:"ip_voting"`
}
type getJamSessionResponseBody jamSessionBody
type setJamSessionResponseBody jamSessionBody

type setJamSessionRequestBody struct {
	Name     string `json:"name"`
	Active   bool   `json:"active"`
	IpVoting bool   `json:"ip_voting"`
}

type playbackBody struct {
	Playback *spotify.PlayerState `json:"playback"`
}
type getPlaybackResponseBody playbackBody
type setPlaybackResponseBody playbackBody

type setPlayBackRequestBody struct {
	Playing  bool       `json:"playing"`
	DeviceID spotify.ID `json:"device_id"`
}

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
		Label:    jamSession.Label,
		Name:     jamSession.Name,
		Active:   jamSession.Active,
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

	jamSession.IpVoteEnabled = body.IpVoting
	jamSession.Active = body.Active
	jamSession.Name = body.Name

	res := setJamSessionResponseBody{
		Label:    jamSession.Label,
		Name:     jamSession.Name,
		Active:   jamSession.Active,
		DeviceID: jamSession.DeviceID,
		IpVoting: jamSession.IpVoteEnabled,
	}

	utils.EncodeJSONBody(w, res)
}

func getPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	res := getPlaybackResponseBody{
		Playback: jamSession.PlaybackState,
	}

	utils.EncodeJSONBody(w, res)
}

func setPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body setPlayBackRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession.SetClientID(body.DeviceID)
	jamSession.SetJamSessionState(body.Playing)

	res := setPlaybackResponseBody{
		Playback: jamSession.PlaybackState,
	}

	utils.EncodeJSONBody(w, res)
}

func createJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	loggedIn, err := LoggedIntoSpotify(session)
	if err != nil || !loggedIn {
		http.Error(w, "User Error: Not logged in to spotify", http.StatusUnauthorized)
		log.Printf("@%s User Error: Not logged in to spotify ", session.ID)
		return
	}

	if session.Values[models.SessionLabelTypeKey] != nil {
		if jamSession := GetJamSession(session.Values[models.SessionLabelTypeKey].(string)); jamSession != nil {
			http.Error(w, "JamSession error: User already joined a JamSession", http.StatusUnprocessableEntity)
			return
		}
	}

	token, err := utils.ParseTokenFromSession(session)
	if err != nil {
		http.Error(w, "User Error: failed to parse token", http.StatusUnauthorized)
		log.Printf("@%s User Error: failed to parse token", session.ID)
		return
	}

	if !token.Valid() {
		http.Error(w, "User Error: token not valid", http.StatusUnauthorized)
		log.Printf("@%s User Error: token not valid", session.ID)
		return
	}

	client := spotifyAuthenticator.NewClient(token)

	label, err := GenerateNewJamSession(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't create jamSession: %s", session.ID, err.Error())
	}

	session.Values[models.SessionLabelTypeKey] = label
	session.Values[models.SessionUserTypeKey] = models.UserTypeHost
	SaveSession(w, r, session)

	res := createJamSessionResponseBody{Label: label}
	utils.EncodeJSONBody(w, res)
}

func joinJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if session.Values[models.SessionLabelTypeKey] != nil {
		if jamSession := GetJamSession(session.Values[models.SessionLabelTypeKey].(string)); jamSession != nil {
			http.Error(w, "JamSession error: User already joined a JamSession", http.StatusUnprocessableEntity)
			return
		}
	}

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

	session.Values[models.SessionUserTypeKey] = models.UserTypeGuest
	session.Values[models.SessionLabelTypeKey] = jamSession.Label
	SaveSession(w, r, session)

	res := joinResponseBody{Label: jamSession.Label}
	utils.EncodeJSONBody(w, res)
}

func leaveJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if LoggedInAsHost(session) {
		label := session.Values[models.SessionLabelTypeKey].(string)
		jamSession := GetJamSession(label)
		if jamSession != nil {
			body := jamSessionStateResponseBody{
				CurrentSong: jamSession.CurrentSong,
				State:       jamSession.PlaybackState,
			}
			Socket.BroadcastToRoom(SocketNamespace, jamSession.Label, SocketEventPlayback, body)
			Socket.BroadcastToRoom(SocketNamespace, jamSession.Label, SocketEventClose, CloseTypeHostLeft)
			DeleteJamSession(label)
		}
	}

	session.Values[models.SessionUserTypeKey] = models.UserTypeNew
	session.Values[models.SessionLabelTypeKey] = nil
	SaveSession(w, r, session)

	res := leaveJamSessionResponseBody{Success: true}
	utils.EncodeJSONBody(w, res)
}
