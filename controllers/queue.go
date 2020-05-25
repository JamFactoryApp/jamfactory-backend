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
	party := utils.PartyFromRequestContext(r)

	voteID := session.ID
	if party.IpVoteEnabled {
		voteID = r.RemoteAddr
	}

	queue := party.Queue.GetObjectWithoutId(voteID)
	utils.EncodeJSONBody(w, queue)
}

func addPlaylist(w http.ResponseWriter, r *http.Request) {
	party := utils.PartyFromRequestContext(r)

	var body addPlaylistRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	playlist, err := party.Client.GetPlaylistTracks(body.PlaylistURI)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Debug("Could not get playlist: ", err.Error())
		return
	}

	for i := 0; i < len(playlist.Tracks); i++ {
		party.Queue.Vote(models.UserTypeHost, playlist.Tracks[i].Track)
	}

	queue := party.Queue.GetObjectWithoutId("")
	Socket.BroadcastToRoom("/", party.Label, SocketEventQueue, party.Queue.GetObjectWithoutId(""))
	utils.EncodeJSONBody(w, queue)
}

func vote(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	party := utils.PartyFromRequestContext(r)

	var body voteRequestBody
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	voteID := session.ID
	if party.IpVoteEnabled {
		voteID = r.RemoteAddr
	}

	party.Queue.Vote(voteID, body.Song)
	queue := party.Queue.GetObjectWithoutId(voteID)
	Socket.BroadcastToRoom("/", party.Label, SocketEventQueue, party.Queue.GetObjectWithoutId(""))
	utils.EncodeJSONBody(w, queue)
}
