package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"jamfactory-backend/utils"
	"net/http"
)

type voteRequestBody struct {
	Song spotify.FullTrack `json:"song"`
}

type addPlaylistRequestBody struct {
	PlaylistURI spotify.ID `json:"uri"`
}

func getQueue(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	jamSession := utils.JamSessionFromRequestContext(r)

	voteID := session.ID
	if jamSession.IpVoteEnabled {
		voteID = r.RemoteAddr
	}

	queue := jamSession.Queue.GetObjectWithoutId(voteID)
	utils.EncodeJSONBody(w, queue)
}

func addPlaylist(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body addPlaylistRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	playlist, err := jamSession.Client.GetPlaylistTracks(body.PlaylistURI)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Debug("Could not get playlist: ", err.Error())
		return
	}

	for i := 0; i < len(playlist.Tracks); i++ {
		jamSession.Queue.Vote(models.UserTypeHost, playlist.Tracks[i].Track)
	}

	queue := jamSession.Queue.GetObjectWithoutId("")
	Socket.BroadcastToRoom("/", jamSession.Label, SocketEventQueue, jamSession.Queue.GetObjectWithoutId(""))
	utils.EncodeJSONBody(w, queue)
}

func vote(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	jamSession := utils.JamSessionFromRequestContext(r)

	var body voteRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	voteID := session.ID
	if jamSession.IpVoteEnabled {
		voteID = r.RemoteAddr
	}

	jamSession.Queue.Vote(voteID, body.Song)
	queue := jamSession.Queue.GetObjectWithoutId(voteID)
	Socket.BroadcastToRoom("/", jamSession.Label, SocketEventQueue, jamSession.Queue.GetObjectWithoutId(""))
	utils.EncodeJSONBody(w, queue)
}
