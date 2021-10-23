package server

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

func (s *Server) getQueue(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	voteID := s.CurrentVoteID(r)
	tracks := jamSession.Queue().For(voteID)

	utils.EncodeJSONBody(w, types.GetQueueResponse{
		Tracks: tracks,
	})
}

func (s *Server) getQueueHistory(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	voteID := s.CurrentVoteID(r)
	history := jamSession.Queue().GetHistory(voteID)

	utils.EncodeJSONBody(w, types.GetQueueHistoryResponse{
		History: history,
	})
}

func (s *Server) exportQueue(w http.ResponseWriter, r *http.Request) {
	var body types.PutQueueExportRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := s.CurrentJamSession(r)
	voteID := s.CurrentVoteID(r)
	tracks := make([]types.Song, 0)
	if body.IncludeHistory {
		tracks = append(tracks, jamSession.Queue().GetHistory(voteID)...)
	}

	if body.IncludeQueue {
		tracks = append(tracks, jamSession.Queue().For(voteID)...)
	}
	if len(tracks) == 0 {
		s.errBadRequest(w, errors.New("No songs to export"), log.DebugLevel)
		return
	}

	ids := make([]spotify.ID, len(tracks))
	for i := range tracks {
		ids[i] = tracks[i].Song.ID
	}
	desc := jamSession.Name() + "  exported queue at " + time.Now().Format("02.01.2006, 15:01") + ". https://jamfactory.app"
	err := jamSession.CreatePlaylist(body.PlaylistName, desc, ids)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.SuccessResponse{
		Success: true,
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
