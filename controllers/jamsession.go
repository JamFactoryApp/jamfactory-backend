package controllers

import (
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/notifications"
	"github.com/jamfactoryapp/jamfactory-backend/types"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"net/http"
	"strings"
)

func getJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	res := types.GetJamResponse{
		Label:      jamSession.Label,
		Name:       jamSession.Name,
		Active:     jamSession.Active,
		VotingType: jamSession.VotingType,
	}

	utils.EncodeJSONBody(w, res)
}

func setJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body types.PutJamRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	if body.VotingType.Set && body.VotingType.Valid {
		switch body.VotingType.Value {
		case types.SessionVotingType:
			jamSession.VotingType = types.SessionVotingType
		case types.IpVotingType:
			jamSession.VotingType = types.IpVotingType
		default:
			http.Error(w, "Not supported voting type", http.StatusUnprocessableEntity)
			return
		}
	}

	if body.Active.Set && body.Active.Valid {
		jamSession.Active = body.Active.Value
	}

	if body.Name.Set && body.Name.Valid {
		jamSession.Name = body.Name.Value
	}

	res := types.PutJamResponse{
		Label:      jamSession.Label,
		Name:       jamSession.Name,
		Active:     jamSession.Active,
		VotingType: jamSession.VotingType,
	}

	utils.EncodeJSONBody(w, res)
}

func getPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)
	jamSession.Lock()
	defer jamSession.Unlock()

	res := types.GetJamPlaybackResponse{
		Playback: *jamSession.PlaybackState,
		DeviceID: jamSession.DeviceID,
	}

	utils.EncodeJSONBody(w, res)
}

func setPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body types.PutJamPlaybackRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	if body.DeviceID.Set && body.DeviceID.Value != "" {
		jamSession.SetClientID(spotify.ID(body.DeviceID.Value))
		log.Debug("Set ID")
	}

	if body.Playing.Set {
		if jamSession.DeviceID != "" {
			jamSession.SetJamSessionState(body.Playing.Value)
			log.Debug("Set State")
		} else {
			http.Error(w, "User Error: No Playback Device Selected", http.StatusForbidden)
			log.Debug("User Error: No Playback Device Selected")
			return
		}
	}

	res := types.PutJamPlaybackResponse{
		Playback: *jamSession.PlaybackState,
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

	if session.Values[utils.SessionLabelTypeKey] != nil {
		if jamSession := GetJamSession(session.Values[utils.SessionLabelTypeKey].(string)); jamSession != nil {
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

	jamSession, err := GenerateNewJamSession(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("@%s Couldn't create jamSession: %s", session.ID, err.Error())
		return
	}

	session.Values[utils.SessionLabelTypeKey] = jamSession.Label
	session.Values[utils.SessionUserTypeKey] = models.UserTypeHost
	SaveSession(w, r, session)

	res := types.GetJamCreateResponse{Label: jamSession.Label}
	utils.EncodeJSONBody(w, res)
}

func joinJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if session.Values[utils.SessionLabelTypeKey] != nil {
		if jamSession := GetJamSession(session.Values[utils.SessionLabelTypeKey].(string)); jamSession != nil {
			http.Error(w, "JamSession error: User already joined a JamSession", http.StatusUnprocessableEntity)
			return
		}
	}

	var body types.PutJamJoinRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := GetJamSession(strings.ToUpper(body.Label))

	if jamSession == nil {
		http.Error(w, "JamSession Error: Could not find a jamSession with the submitted label", http.StatusNotFound)
		log.Printf("@%s JamSession Error: Could not find a jamSession with the submitted label", session.ID)
		return
	}

	session.Values[utils.SessionUserTypeKey] = models.UserTypeGuest
	session.Values[utils.SessionLabelTypeKey] = jamSession.Label
	SaveSession(w, r, session)

	res := types.PutJamJoinResponse{Label: jamSession.Label}
	utils.EncodeJSONBody(w, res)
}

func leaveJamSession(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)

	if LoggedInAsHost(session) {
		label := session.Values[utils.SessionLabelTypeKey].(string)
		jamSession := GetJamSession(label)
		if jamSession != nil {
			body := types.SocketJamState{
				CurrentSong: jamSession.CurrentSong,
				State:       jamSession.PlaybackState,
			}
			jamSession.NotifyClients(&notifications.Message{
				Event:   notifications.Playback,
				Message: body,
			})
			jamSession.NotifyClients(&notifications.Message{
				Event:   notifications.Close,
				Message: notifications.HostLeft,
			})
			DeleteJamSession(label)
		}
	}

	session.Values[utils.SessionUserTypeKey] = models.UserTypeNew
	session.Values[utils.SessionLabelTypeKey] = nil
	SaveSession(w, r, session)

	res := types.GetJamLeaveResponse{
		Success: true,
	}
	utils.EncodeJSONBody(w, res)
}
