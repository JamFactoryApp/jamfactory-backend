package server

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	"net/http"

	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/users"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/jamsession"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	log "github.com/sirupsen/logrus"
)

func (s *Server) getMemberResponse(members jamsession.Members) types.GetJamMembersResponse {
	memberResponse := make([]types.JamMember, 0)
	for _, member := range members {
		user, err := s.users.Get(member.Identifier())
		if err != nil {
			log.Warn("User for identifier not found", member.Identifier())
			continue
		}
		memberResponse = append(memberResponse, types.JamMember{
			DisplayName: user.UserName,
			Identifier:  user.Identifier,
			Rights:      member.Permissions(),
		})
	}
	return types.GetJamMembersResponse{Members: memberResponse}
}

func (s *Server) getMembers(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	utils.EncodeJSONBody(w, s.getMemberResponse(jamSession.Members()))
}

func (s *Server) setMembers(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamMemberRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)

	// Validate Request
	if len(body.Members) != len(jamSession.Members()) {
		s.errBadRequest(w, apierrors.ErrWrongMemberCount, log.DebugLevel)
		return
	}
	hostCount := 0
	for _, requestMember := range body.Members {
		if jamsession.ContainsPermissions(types.RightHost, requestMember.Rights) {
			hostCount++
		}

		if !jamsession.ValidPermissions(requestMember.Rights) {
			s.errBadRequest(w, apierrors.ErrBadRight, log.DebugLevel)
			return
		}
		included := false
		for _, availableMembers := range jamSession.Members() {
			if requestMember.Identifier == availableMembers.Identifier() {
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
	for _, availableMembers := range jamSession.Members() {
		for _, requestMember := range body.Members {
			if requestMember.Identifier == availableMembers.Identifier() {
				availableMembers.SetPermissions(requestMember.Rights)
			}
		}
	}

	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Members,
		Message: s.getMemberResponse(jamSession.Members()),
	})

	utils.EncodeJSONBody(w, s.getMemberResponse(jamSession.Members()))
}

func (s *Server) getJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	utils.EncodeJSONBody(w, types.GetJamResponse{
		Label:  jamSession.JamLabel(),
		Name:   jamSession.Name(),
		Active: jamSession.Active(),
	})

}

func (s *Server) setJamSession(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)

	if body.Active.Set && body.Active.Valid {
		if body.Active.Value && !jamSession.Active() {
			song, err := jamSession.Queue().GetNext()
			if err != nil {
				s.errBadRequest(w, apierrors.ErrQueueEmpty, log.DebugLevel)
				return
			}
			if jamSession.GetDevice().ID == "" {
				s.errBadRequest(w, apierrors.ErrNoDevice, log.DebugLevel)
				return
			}
			jamSession.Play(jamSession.GetDevice(), song)
			message := types.GetQueueResponse{Tracks: jamSession.Queue().Tracks()}
			jamSession.NotifyClients(&notifications.Message{
				Event:   notifications.Queue,
				Message: message,
			})
		}
		jamSession.SetActive(body.Active.Value)
	}

	if body.Name.Set && body.Name.Valid {
		jamSession.SetName(body.Name.Value)
	}
	jamSession.NotifyClients(&notifications.Message{
		Event: notifications.Jam,
		Message: types.SocketJamMessage{
			Label:  jamSession.JamLabel(),
			Name:   jamSession.Name(),
			Active: jamSession.Active(),
		},
	})
	utils.EncodeJSONBody(w, types.GetJamResponse{
		Label:  jamSession.JamLabel(),
		Name:   jamSession.Name(),
		Active: jamSession.Active(),
	})
}

func (s *Server) getPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	utils.EncodeJSONBody(w, types.GetJamPlaybackResponse{
		Playback: jamSession.GetPlayerState(),
		DeviceID: jamSession.GetDevice().ID,
	})
}

func (s *Server) setPlayback(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamPlaybackRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)

	if body.Playing.Set && body.Playing.Valid {
		if err := jamSession.SetState(body.Playing.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
		playerState := jamSession.GetPlayerState()
		playerState.Playing = body.Playing.Value
		jamSession.SetPlayerState(playerState)
	}

	if body.DeviceID.Set && body.DeviceID.Valid {
		if err := jamSession.SetDevice(body.DeviceID.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
		}
	}

	utils.EncodeJSONBody(w, types.PutJamPlaybackResponse{
		Playback: jamSession.GetPlayerState(),
		DeviceID: jamSession.GetDevice().ID,
	})
}

func (s *Server) createJamSession(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	user, err := s.users.Get(s.CurrentIdentifier(r))

	if !user.SpotifyToken.Valid() {
		s.errForbidden(w, apierrors.ErrTokenInvalid, log.DebugLevel)
		return
	}

	jamSession, err := s.jamFactory.NewJamSession(user)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	go jamSession.Conductor()

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	utils.EncodeJSONBody(w, types.GetJamCreateResponse{
		Label: jamSession.JamLabel(),
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

	//Check if a user for the request already exits
	user, err := s.users.Get(s.CurrentIdentifier(r))
	if err != nil {
		// Create guest user from session
		hash := sha1.Sum([]byte(session.ID))
		identifier := hex.EncodeToString(hash[:])
		username := "Guest " + string([]rune(base32.StdEncoding.EncodeToString(hash[:]))[0:5])
		user = users.New(identifier, username, users.UserTypeSession, nil)
		if err := s.users.Save(user); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	sessions.SetIdentifier(session, user.Identifier)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	jamSession.Members().Add(user.Identifier, []types.Permission{types.RightsGuest})

	jamSession.NotifyClients(&notifications.Message{
		Event:   notifications.Members,
		Message: s.getMemberResponse(jamSession.Members()),
	})

	utils.EncodeJSONBody(w, types.PutJamJoinResponse{
		Label: jamLabel,
	})
}

func (s *Server) leaveJamSession(w http.ResponseWriter, r *http.Request) {

	user := s.CurrentUser(r)
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err == nil {
		member, err := jamSession.Members().Get(user.Identifier)
		if err != nil {
			s.errBadRequest(w, err, log.DebugLevel)
			return
		}
		isHost := member.HasPermissions([]types.Permission{types.RightHost})
		if isHost {
			if jamSession.Members().Remove(user.Identifier) {
				jamSession.NotifyClients(&notifications.Message{
					Event:   notifications.Close,
					Message: notifications.HostLeft,
				})
				if err := s.jamFactory.DeleteJamSession(jamSession.JamLabel()); err != nil {
					s.errInternalServerError(w, err, log.DebugLevel)
					return
				}
			}
		} else {
			jamSession.Members().Remove(user.Identifier)
			jamSession.NotifyClients(&notifications.Message{
				Event:   notifications.Members,
				Message: s.getMemberResponse(jamSession.Members()),
			})
		}
	}

	utils.EncodeJSONBody(w, types.GetJamLeaveResponse{
		Success: true,
	})
}
