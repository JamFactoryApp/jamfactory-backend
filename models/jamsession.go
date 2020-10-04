package models

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/types"
	"sync"
)

type JamSession struct {
	Label         string
	Name          string
	VotingType    types.VotingType
	Active        bool
	Context       context.Context
	Queue         *Queue
	Client        spotify.Client
	DeviceID      spotify.ID
	CurrentSong   *spotify.FullTrack
	PlaybackState *spotify.PlayerState
	mux           sync.Mutex
}

type JamSessions []*JamSession

func (jamSession *JamSession) StartNextSong() {
	log.WithField("JamSession", jamSession.Label).Trace("Model event: Start next song for jamSession")

	song, err := jamSession.Queue.GetNextSong()

	if err != nil {
		log.WithField("JamSession", jamSession.Label).Trace("Model event: Queue is empty")
		return
	}

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{jamSession.CurrentSong.URI},
	}

	err = jamSession.Client.PlayOpt(&playOptions)
	if err != nil {
		log.WithField("JamSession", jamSession.Label).Error("Error starting next song: ", err.Error())
		return
	}

	jamSession.CurrentSong = song.Song
	err = jamSession.Queue.AdvanceQueue()

	if err != nil {
		log.WithField("JamSession", jamSession.Label).Trace("Model event: Could not advance queue")
		return
	}

}

func (jamSession *JamSession) SetJamSessionState(state bool) {
	log.WithField("JamSession", jamSession.Label).Trace("Model event: Set jamSession enabled")

	var fragment string
	var err error

	if state {
		err = jamSession.Client.Play()
		fragment = "play"
	} else {
		err = jamSession.Client.Pause()
		fragment = "pause"
	}

	if err != nil {
		log.WithField("JamSession", jamSession.Label).Warn(fmt.Sprintf("Error setting client to %s\n", fragment))
	}

	jamSession.Active = state
	jamSession.PlaybackState.Playing = state
}

func (jamSession *JamSession) SetClientID(id spotify.ID) {
	if jamSession.DeviceID != id {
		err := jamSession.Client.TransferPlayback(id, jamSession.Active)

		if err != nil {
			log.Debug("Error setting Device ID: ", err.Error())
		} else {
			jamSession.DeviceID = id
		}
	}
}

func (jamSession *JamSession) UpdatePlaybackState(state *spotify.PlayerState) {
	jamSession.Lock()
	defer jamSession.Unlock()
	jamSession.PlaybackState = state
}

func (jamSession *JamSession) UpdateCurrentSong(item *spotify.FullTrack) {
	jamSession.Lock()
	defer jamSession.Unlock()
	jamSession.CurrentSong = item
}

func (jamSession *JamSession) Lock() {
	jamSession.mux.Lock()
}

func (jamSession *JamSession) Unlock() {
	jamSession.mux.Unlock()
}
