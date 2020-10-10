package controllers

import (
	"github.com/jamfactoryapp/jamfactory-backend/models"
	"github.com/jamfactoryapp/jamfactory-backend/notifications"
	"github.com/jamfactoryapp/jamfactory-backend/types"
	"github.com/jamfactoryapp/jamfactory-backend/utils"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
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

func addCollection(w http.ResponseWriter, r *http.Request) {
	session := utils.SessionFromRequestContext(r)
	jamSession := utils.JamSessionFromRequestContext(r)

	var body types.PutQueueCollectionRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	switch body.CollectionType {
	case "playlist":
		playlist, err := jamSession.Client.GetPlaylistTracks(spotify.ID(body.CollectionID))

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			log.Debug("Could not get playlist: ", err.Error())
			return
		}

		for i := 0; i < len(playlist.Tracks); i++ {
			jamSession.Queue.Vote(models.UserTypeHost, &playlist.Tracks[i].Track)
		}

	case "album":
		album, err := jamSession.Client.GetAlbumTracks(spotify.ID(body.CollectionID))

		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			log.Debug("Could not get album: ", err.Error())
			return
		}

		ids := make([]spotify.ID, len(album.Tracks))
		for i := 0; i < len(album.Tracks); i++ {
			ids[i] = album.Tracks[i].ID
		}

		tracks, err := jamSession.Client.GetTracks(ids...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			log.Debug("Error getting album tracks ", err.Error())
			return
		}

		for i := 0; i < len(tracks); i++ {
			jamSession.Queue.Vote(models.UserTypeHost, tracks[i])
		}

	default:
		http.Error(w, "Unsupported collection type", http.StatusUnprocessableEntity)
		return
	}

	voteID := session.ID
	if jamSession.VotingType == types.IpVotingType {
		voteID = r.RemoteAddr
	}

	message := types.PutQueuePlaylistsResponse{
		Queue: jamSession.Queue.GetObjectWithoutId(""),
	}
	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Queue,
		Message: message,
	})

	res := types.PutQueuePlaylistsResponse{
		Queue: jamSession.Queue.GetObjectWithoutId(voteID),
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

	message := types.PutQueueVoteResponse{
		Queue: jamSession.Queue.GetObjectWithoutId(""),
	}
	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Queue,
		Message: message,
	})

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
	if jamSession.VotingType == types.IpVotingType {
		voteID = r.RemoteAddr
	}

	if ok := jamSession.Queue.DeleteSong(spotify.ID(body.TrackID)); !ok {
		http.Error(w, "Could not find song", http.StatusUnprocessableEntity)
		return
	}

	message := types.PutQueuePlaylistsResponse{
		Queue: jamSession.Queue.GetObjectWithoutId(""),
	}
	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Queue,
		Message: message,
	})

	res := types.DeleteQueueSongResponse{
		Queue: jamSession.Queue.GetObjectWithoutId(voteID),
	}
	utils.EncodeJSONBody(w, res)
}
