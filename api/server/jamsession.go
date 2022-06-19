package server

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/permissions"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	"github.com/zmb3/spotify"
	"net/http"

	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	log "github.com/sirupsen/logrus"
)

func (s *Server) getMemberResponse(members jamsession.Members) types.GetJamMembersResponse {
	memberResponse := make([]types.JamMember, 0)
	for _, member := range members {
		user, err := s.users.GetUserByIdentifier(member.GetIdentifier())
		if err != nil {
			log.Warn("User for identifier not found", member.GetIdentifier())
			continue
		}
		userInfo, err := user.GetInfo()
		if err != nil {
			log.Warn("UserInfo for identifier not found", member.GetIdentifier())
			continue
		}
		memberResponse = append(memberResponse, types.JamMember{
			DisplayName: userInfo.UserName,
			Identifier:  user.Identifier,
			Permissions: member.GetPermissions(),
		})
	}
	return types.GetJamMembersResponse{Members: memberResponse}
}

func (s *Server) getMembers(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	utils.EncodeJSONBody(w, s.getMemberResponse(*members))
}

func (s *Server) setMembers(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamMemberRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	// Validate Request
	if len(body.Members) != len(*members) {
		s.errBadRequest(w, apierrors.ErrWrongMemberCount, log.DebugLevel)
		return
	}
	hostCount := 0
	for _, requestMember := range body.Members {
		if !requestMember.Permissions.Valid() {
			s.errBadRequest(w, apierrors.ErrBadRight, log.DebugLevel)
			return
		}

		member := jamsession.NewMember(requestMember.Identifier, requestMember.Permissions...)
		if member.HasPermissions(permissions.Host) {
			hostCount++
		}

		included := false
		for _, availableMembers := range *members {
			if requestMember.Identifier == availableMembers.GetIdentifier() {
				included = true
				break
			}
		}
		if !included {
			s.errBadRequest(w, apierrors.ErrMissingMember, log.DebugLevel)
			return
		}
	}
	if hostCount != 1 {
		s.errBadRequest(w, apierrors.ErrOnlyOneHost, log.DebugLevel)
		return
	}
	// Request is valid. Apply changes
	for _, availableMembers := range *members {
		for _, requestMember := range body.Members {
			if requestMember.Identifier == availableMembers.GetIdentifier() {
				availableMembers.SetPermissions(requestMember.Permissions...)
			}
		}
	}

	if err := jamSession.SetMembers(members); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Members,
		Message: s.getMemberResponse(*members),
	})

	utils.EncodeJSONBody(w, s.getMemberResponse(*members))
}

func (s *Server) getJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	settings, err := jamSession.GetSettings()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	utils.EncodeJSONBody(w, types.GetJamResponse{
		Label:  jamSession.JamLabel,
		Name:   settings.Name,
		Active: settings.Active,
	})
}

func (s *Server) setJamSession(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)
	settings, err := jamSession.GetSettings()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	currentQueue, err := jamSession.GetQueue()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	host, err := s.users.GetUserByIdentifier(members.Host().Identifier)
	if err != nil {
		s.errInternalServerError(w, apierrors.ErrMissingMember, log.WarnLevel)
		return
	}

	if body.Active.Set && body.Active.Valid {
		if body.Active.Value && !settings.Active {
			song, err := currentQueue.GetNext()
			if err != nil {
				s.errBadRequest(w, apierrors.ErrQueueEmpty, log.DebugLevel)
				return
			}
			if host.GetPlayerState().Device.ID == "" {
				s.errBadRequest(w, apierrors.ErrNoDevice, log.DebugLevel)
				return
			}

			// TODO: jamSession.Play or user.Play ???
			jamSession.Play(song.Track, true)
			jamSession.SocketQueueUpdate()
		}
		settings.Active = body.Active.Value
	}

	if body.Name.Set && body.Name.Valid {
		settings.Name = body.Name.Value
	}

	if body.Password.Set && body.Password.Valid {
		settings.Password = body.Password.Value
	}

	if err := jamSession.SetSettings(settings); err != nil {
		s.errInternalServerError(w, apierrors.ErrMissingMember, log.WarnLevel)
		return
	}

	jamSession.NotifyClients(&notifications.Message{
		Event: notifications.Jam,
		Message: types.SocketJamMessage{
			Label:  jamSession.JamLabel,
			Name:   settings.Name,
			Active: settings.Active,
		},
	})
	utils.EncodeJSONBody(w, types.GetJamResponse{
		Label:  jamSession.JamLabel,
		Name:   settings.Name,
		Active: settings.Active,
	})
}

func (s *Server) getPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	host, err := s.users.GetUserByIdentifier(members.Host().Identifier)
	if err != nil {
		s.errInternalServerError(w, apierrors.ErrMissingMember, log.WarnLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetPlaybackResponse{
		Playback: host.GetPlayerState(),
		DeviceID: host.GetPlayerState().Device.ID,
	})
}

func (s *Server) setPlayback(w http.ResponseWriter, r *http.Request) {
	var body types.PutPlaybackRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	host, err := s.users.GetUserByIdentifier(members.Host().Identifier)
	if err != nil {
		s.errInternalServerError(w, apierrors.ErrMissingMember, log.WarnLevel)
		return
	}

	if body.Playing.Set && body.Playing.Valid {
		if err := host.SetState(body.Playing.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
		playerState := host.GetPlayerState()
		playerState.Playing = body.Playing.Value
		host.SetPlayerState(playerState)
	}

	if body.Volume.Set && body.Volume.Valid {
		if err := host.SetVolume(body.Volume.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	if body.DeviceID.Set && body.DeviceID.Valid {
		if err := host.SetDevice(body.DeviceID.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	utils.EncodeJSONBody(w, types.PutJamPlaybackResponse{
		Playback: host.GetPlayerState(),
		DeviceID: host.GetPlayerState().Device.ID,
	})
}

func (s *Server) playSong(w http.ResponseWriter, r *http.Request) {
	var body types.PutPlaySongRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	host, err := s.users.GetUserByIdentifier(members.Host().Identifier)
	track, err := host.GetTrack(body.TrackID)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	if err := jamSession.Play(track, body.Remove); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.SuccessResponse{
		Success: true,
	})
}

func (s *Server) createJamSession(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	userInfo, err := user.GetInfo()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	if !userInfo.SpotifyToken.Valid() {
		s.errForbidden(w, apierrors.ErrTokenInvalid, log.DebugLevel)
		return
	}

	jamSession, err := s.jamFactory.NewJamSession(user)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetJamCreateResponse{
		Label: jamSession.JamLabel,
	})
}

func (s *Server) joinJamSession(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamJoinRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		return
	}

	session := s.CurrentSession(r)
	jamLabel := body.Label

	jamSession, err := s.jamFactory.GetJamSessionByLabel(jamLabel)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	settings, err := jamSession.GetSettings()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}
	members, err := jamSession.GetMembers()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	// Check if the password is correct
	if body.Password != settings.Password {
		s.errUnauthorized(w, apierrors.ErrWrongPassword, log.DebugLevel)
		return
	}

	//Check if a user for the request already exits
	user, err := s.users.GetUserByIdentifier(s.CurrentIdentifier(r))
	if err != nil {
		// Create guest user from session
		hash := sha1.Sum([]byte(session.ID))
		identifier := hex.EncodeToString(hash[:])
		username := "Guest " + string([]rune(base32.StdEncoding.EncodeToString(hash[:]))[0:5])
		user, err = s.users.NewUser(identifier, username, users.UserTypeSession, nil)
		if err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	sessions.SetIdentifier(session, user.Identifier)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	members.Add(user.Identifier, permissions.Guest)

	if err := jamSession.SetMembers(members); err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Members,
		Message: s.getMemberResponse(*members),
	})

	utils.EncodeJSONBody(w, types.PutJamJoinResponse{
		Label: jamLabel,
	})
}

func (s *Server) leaveJamSession(w http.ResponseWriter, r *http.Request) {

	user := s.CurrentUser(r)
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err == nil {
		members, err := jamSession.GetMembers()
		if err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
		member, err := members.Get(user.Identifier)
		if err != nil {
			s.errBadRequest(w, err, log.DebugLevel)
			return
		}
		isHost := member.HasPermissions(permissions.Host)
		if isHost {
			if members.Remove(user.Identifier) {
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.HostLeft,
				})
				if err := s.jamFactory.DeleteJamSession(jamSession.JamLabel); err != nil {
					s.errInternalServerError(w, err, log.DebugLevel)
					return
				}
			}
		} else {
			members.Remove(user.Identifier)
			jamSession.NotifyClients(&notifications.Message{
				Event:   notifications.Members,
				Message: s.getMemberResponse(*members),
			})
		}
		if err := jamSession.SetMembers(members); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	utils.EncodeJSONBody(w, types.GetJamLeaveResponse{
		Success: true,
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
