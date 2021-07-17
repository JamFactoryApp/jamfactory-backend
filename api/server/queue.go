package server

import (
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) getQueue(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	voteID := s.CurrentVoteID(r)
	tracks := jamSession.Queue().For(voteID)

	utils.EncodeJSONBody(w, types.GetQueueResponse{
		Tracks: tracks,
	})
}

func (s *Server) addCollection(w http.ResponseWriter, r *http.Request) {
	var body types.PutQueueCollectionRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := s.CurrentJamSession(r)

	err := jamSession.AddCollection(body.CollectionType, body.CollectionID)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	voteID := s.CurrentVoteID(r)
	tracks := jamSession.Queue().For(voteID)

	utils.EncodeJSONBody(w, types.PutQueuePlaylistsResponse{
		Tracks: tracks,
	})
}

func (s *Server) vote(w http.ResponseWriter, r *http.Request) {
	var body types.PutQueueVoteRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := s.CurrentJamSession(r)
	voteID := s.CurrentVoteID(r)

	if err := jamSession.Vote(body.TrackID, voteID); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	tracks := jamSession.Queue().For(voteID)

	utils.EncodeJSONBody(w, types.PutQueueVoteResponse{
		Tracks: tracks,
	})
}

func (s *Server) deleteSong(w http.ResponseWriter, r *http.Request) {
	var body types.DeleteQueueSongRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := s.CurrentJamSession(r)
	voteID := s.CurrentVoteID(r)

	if err := jamSession.DeleteSong(body.TrackID); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	tracks := jamSession.Queue().For(voteID)

	utils.EncodeJSONBody(w, types.DeleteQueueSongResponse{
		Tracks: tracks,
	})
}
