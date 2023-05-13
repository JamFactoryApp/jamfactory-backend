package server

import (
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) userMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identifier := s.CurrentIdentifier(r)
		user, err := s.users.GetUserByIdentifier(r.Context(), identifier)
		if err != nil {
			user = users.NewEmpty()
		}
		ctx := users.NewContext(r.Context(), user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	userInfo, err := user.GetInfo()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	var jamLabel string
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err != nil {
		jamLabel = ""
	} else {
		jamLabel = jamSession.JamLabel
	}

	spotifyAuthorized := false
	if userInfo.SpotifyToken != nil && userInfo.SpotifyToken.Valid() {
		spotifyAuthorized = true
	}

	utils.EncodeJSONBody(w, types.GetUserResponse{
		Identifier:        user.Identifier,
		DisplayName:       userInfo.UserName,
		UserType:          string(userInfo.UserType),
		StartListen:       userInfo.UserStartListening,
		JoinedLabel:       jamLabel,
		SpotifyAuthorized: spotifyAuthorized,
	})
}

func (s *Server) setUser(w http.ResponseWriter, r *http.Request) {
	var body types.PutUserRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}
	user := s.CurrentUser(r)
	userInfo, err := user.GetInfo()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	userInfo.UserName = body.DisplayName
	userInfo.UserStartListening = body.StartListen
	user.SetInfo(userInfo)

	var jamLabel string
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err != nil {
		jamLabel = ""
	} else {
		jamLabel = jamSession.JamLabel
	}

	spotifyAuthorized := false
	if userInfo.SpotifyToken != nil && userInfo.SpotifyToken.Valid() {
		spotifyAuthorized = true
	}

	utils.EncodeJSONBody(w, types.GetUserResponse{
		Identifier:        user.Identifier,
		DisplayName:       userInfo.UserName,
		UserType:          string(userInfo.UserType),
		JoinedLabel:       jamLabel,
		StartListen:       userInfo.UserStartListening,
		SpotifyAuthorized: spotifyAuthorized,
	})
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	s.users.DeleteUser(r.Context(), user.Identifier)
	// TODO: Make sure that user is not used anywhere
	utils.EncodeJSONBody(w, types.DeleteUserResponse{
		Success: true,
	})
}

func (s *Server) getUserPlayback(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	userInfo, err := user.GetInfo()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	// Getting the playback only makes sense for Spotify Users
	if !(userInfo.UserType == users.UserTypeSpotify) {
		s.errUnauthorized(w, apierrors.ErrUserTypeInvalid, log.TraceLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetPlaybackResponse{
		Playback: user.GetPlayerState(),
		DeviceID: user.GetPlayerState().Device.ID,
	})
}

func (s *Server) setUserPlayback(w http.ResponseWriter, r *http.Request) {
	var body types.PutPlaybackRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	user := s.CurrentUser(r)
	userInfo, err := user.GetInfo()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	// Setting the playback only makes sense for Spotify Users
	if !(userInfo.UserType == users.UserTypeSpotify) {
		s.errUnauthorized(w, apierrors.ErrUserTypeInvalid, log.TraceLevel)
		return
	}

	if body.Playing.Set && body.Playing.Valid {
		if err := user.SetState(r.Context(), body.Playing.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
		playerState := user.GetPlayerState()
		playerState.Playing = body.Playing.Value
		user.SetPlayerState(playerState)
	}

	if body.Volume.Set && body.Volume.Valid {
		if err := user.SetVolume(r.Context(), body.Volume.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	if body.DeviceID.Set && body.DeviceID.Valid {
		if err := user.SetDevice(r.Context(), body.DeviceID.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	utils.EncodeJSONBody(w, types.GetPlaybackResponse{
		Playback: user.GetPlayerState(),
		DeviceID: user.GetPlayerState().Device.ID,
	})
}

func (s *Server) getUserPlaylists(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)

	playlists, err := user.Playlists(r.Context())
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetSpotifyPlaylistsResponse{
		Playlists: playlists,
	})
}

func (s *Server) getUserDevices(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)

	devices, err := user.Devices(r.Context())
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetSpotifyDevicesResponse{
		Devices: devices,
	})
}
