package server

import (
	apierrors "github.com/jamfactoryapp/jamfactory-backend/api/errors"
	"github.com/jamfactoryapp/jamfactory-backend/api/sessions"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/api/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Server) getJamSession(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	utils.EncodeJSONBody(w, types.GetJamResponse{
		Label:      jamSession.JamLabel(),
		Name:       jamSession.Name(),
		Active:     jamSession.Active(),
		VotingType: jamSession.VotingType(),
	})
}

func (s *Server) setJamSession(w http.ResponseWriter, r *http.Request) {
	var body types.PutJamRequest
	if err := utils.DecodeJSONBody(w, r, &body); err != nil {
		s.errBadRequest(w, err, log.DebugLevel)
		return
	}

	jamSession := s.CurrentJamSession(r)

	if body.VotingType.Set && body.VotingType.Valid {
		if err := jamSession.SetVotingType(body.VotingType.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	if body.Active.Set && body.Active.Valid {
		jamSession.SetActive(body.Active.Value)
	}

	if body.Name.Set && body.Name.Valid {
		jamSession.SetName(body.Name.Value)
	}

	utils.EncodeJSONBody(w, types.PutJamResponse{
		Active:     jamSession.Active(),
		Label:      jamSession.JamLabel(),
		Name:       jamSession.Name(),
		VotingType: jamSession.VotingType(),
	})
}

func (s *Server) getPlayback(w http.ResponseWriter, r *http.Request) {
	jamSession := s.CurrentJamSession(r)

	playerState, err := jamSession.PlayerState()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	deviceID, err := jamSession.DeviceID()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	utils.EncodeJSONBody(w, types.GetJamPlaybackResponse{
		Playback: playerState,
		DeviceID: deviceID,
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
	}

	if body.DeviceID.Set && body.DeviceID.Valid {
		if err := jamSession.SetDevice(body.DeviceID.Value); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
		}
	}

	playerState, err := jamSession.PlayerState()
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
	}

	utils.EncodeJSONBody(w, types.PutJamPlaybackResponse{
		Playback: playerState,
	})
}

func (s *Server) createJamSession(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)

	userType := s.CurrentUserType(r)
	if userType == "" {
		s.errBadRequest(w, apierrors.ErrUserTypeInvalid, log.DebugLevel)
	}

	token := s.CurrentToken(r)
	if !token.Valid() {
		s.errForbidden(w, apierrors.ErrTokenInvalid, log.DebugLevel)
		return
	}

	jamSession, err := s.jamFactory.NewJamSession(token)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	go jamSession.Conductor()

	sessions.SetJamLabel(session, jamSession.JamLabel())
	sessions.SetUserType(session, types.UserTypeHost)

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

	_, err := s.jamFactory.GetJamSession(jamLabel)
	if err != nil {
		s.errInternalServerError(w, err, log.DebugLevel)
		return
	}

	sessions.SetUserType(session, types.UserTypeGuest)
	sessions.SetJamLabel(session, jamLabel)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	utils.EncodeJSONBody(w, types.PutJamJoinResponse{
		Label: jamLabel,
	})
}

func (s *Server) leaveJamSession(w http.ResponseWriter, r *http.Request) {
	session := s.CurrentSession(r)
	userType := s.CurrentUserType(r)

	if userType == types.UserTypeHost {
		jamSession := s.CurrentJamSession(r)
		jamSession.SetState(false)
		if err := s.jamFactory.DeleteJamSession(jamSession.JamLabel()); err != nil {
			s.errInternalServerError(w, err, log.DebugLevel)
			return
		}
	}

	sessions.SetJamLabel(session, "")
	sessions.SetUserType(session, types.UserTypeNew)

	if err := session.Save(r, w); err != nil {
		s.errSessionSave(w, err)
		return
	}

	utils.EncodeJSONBody(w, types.GetJamLeaveResponse{
		Success: true,
	})
}
