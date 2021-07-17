package server

import (
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"net/http"
)

func (s *Server) devices(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	devices, err := jamSession.Devices()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetSpotifyDevicesResponse{
		Devices: devices,
	})
}

func (s *Server) playlist(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	playlists, err := jamSession.Playlists()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetSpotifyPlaylistsResponse{
		Playlists: playlists,
	})
}

func (s *Server) search(w http.ResponseWriter, r *http.Request) {
	var body types.PutSpotifySearchRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	jamSession := s.CurrentJamSession(r)

	entry, err := s.jamFactory.Search(jamSession, body.SearchType, body.SearchText)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	searchResult, ok := entry.(*spotify.SearchResult)
	if !ok {
		s.errInternalServerError(w, err, log.ErrorLevel)
		return
	}

	utils.EncodeJSONBody(w, types.PutSpotifySearchResponse{
		Artists:   searchResult.Artists,
		Albums:    searchResult.Albums,
		Playlists: searchResult.Playlists,
		Tracks:    searchResult.Tracks,
	})
}
