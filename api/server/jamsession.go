package server

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/hex"
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) getJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)
	members := jamSession.Members()
	log.Info(len(members))
	memberRespone := make([]types.JamMemberResponse, 0)
	for _, member := range members {
		memberRespone = append(memberRespone, types.JamMemberResponse{
			DisplayName: member.User.UserName,
			Rights:      member.Rights,
		})
	}

	utils.EncodeJSONBody(w, types.GetJamResponse{
		Label:   jamSession.JamLabel(),
		Name:    jamSession.Name(),
		Members: memberRespone,
		Active:  jamSession.Active(),
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
	utils.EncodeJSONBody(w, types.PutJamResponse{
		Active: jamSession.Active(),
		Label:  jamSession.JamLabel(),
		Name:   jamSession.Name(),
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

	if !user.Token.Valid() {
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
		user = s.users.New(identifier, username, types.UserTypeSession, nil)
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

	jamSession.Members().Add(user, []types.MemberRights{types.RightsGuest})

	utils.EncodeJSONBody(w, types.PutJamJoinResponse{
		Label: jamLabel,
	})
}

func (s *Server) leaveJamSession(w http.ResponseWriter, r *http.Request) {

	user := s.CurrentUser(r)
	if jamSession, err := s.jamFactory.GetJamSessionByUser(user); err == nil {
		member, err := jamSession.Members().Get(user)
		if err == nil && member.Has([]types.MemberRights{types.RightHost}) {
			if jamSession.Members().Remove(user) {
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
			jamSession.Members().Remove(user)
		}
	}

	utils.EncodeJSONBody(w, types.GetJamLeaveResponse{
		Success: true,
	})
}
