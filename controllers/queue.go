package controllers

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"jamfactory-backend/types"
	"jamfactory-backend/utils"
	"net/http"
)

func getQueue(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	jamSession := utils.JamSessionFromRequestContext(r)

	voteID := session.ID
	if jamSession.VotingType == types.IpVotingType {
		voteID = r.RemoteAddr
	}

	queue := jamSession.Queue.GetObjectWithoutId(voteID)
	res := types.GetQueueResponse{
		Queue: queue,
	}
	utils.EncodeJSONBody(w, res)
}

func addPlaylist(w http.ResponseWriter, r *http.Request) {
	jamSession := utils.JamSessionFromRequestContext(r)

	var body types.PutQueuePlaylistRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	playlist, err := jamSession.Client.GetPlaylistTracks(spotify.ID(body.PlaylistID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		log.Debug("Could not get playlist: ", err.Error())
		return
	}

	for i := 0; i < len(playlist.Tracks); i++ {
		jamSession.Queue.Vote(models.UserTypeHost, &playlist.Tracks[i].Track)
	}

	queue := jamSession.Queue.GetObjectWithoutId("")
	Socket.BroadcastToRoom(SocketNamespace, jamSession.Label, SocketEventQueue, jamSession.Queue.GetObjectWithoutId(""))

	res := types.PutQueuePlaylistsResponse{
		Queue: queue,
	}
	utils.EncodeJSONBody(w, res)
}

func vote(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	jamSession := utils.JamSessionFromRequestContext(r)

	var body types.PutQueueVoteRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	song, err := jamSession.Client.GetTrack(spotify.ID(body.TrackID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		log.WithField("ID", body.TrackID).Debug("Could not find track: ", err.Error())
		return
	}

	voteID := session.ID
	if jamSession.VotingType == types.IpVotingType {
		voteID = r.RemoteAddr
	}

	jamSession.Queue.Vote(voteID, song)
	queue := jamSession.Queue.GetObjectWithoutId(voteID)
	Socket.BroadcastToRoom(SocketNamespace, jamSession.Label, SocketEventQueue, jamSession.Queue.GetObjectWithoutId(""))

	res := types.PutQueueVoteResponse{
		Queue: queue,
	}
	utils.EncodeJSONBody(w, res)
}

func deleteSong(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	jamSession := utils.JamSessionFromRequestContext(r)

	var body types.DeleteQueueSongRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	voteID := session.ID
	if jamSession.VotingType == types.IpVotingType{
		voteID = r.RemoteAddr
	}

	if ok := jamSession.Queue.DeleteSong(spotify.ID(body.TrackID)); !ok {
		http.Error(w, "Could not find song", http.StatusUnprocessableEntity)
		return
	}

	res := types.DeleteQueueSongResponse{
		Queue: jamSession.Queue.GetObjectWithoutId(voteID),
	}
	Socket.BroadcastToRoom(SocketNamespace, jamSession.Label, SocketEventQueue, jamSession.Queue.GetObjectWithoutId(""))
	utils.EncodeJSONBody(w, res)
}
