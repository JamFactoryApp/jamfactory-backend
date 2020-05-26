package models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"strings"
)

type JamSession struct {
	Label         string
	Queue         *Queue
	IpVoteEnabled bool
	Client        spotify.Client
	DeviceID      spotify.ID
	CurrentSong   *spotify.FullTrack
	PlaybackState *spotify.PlayerState
	User          *spotify.PrivateUser
	Active        bool
}

type JamSessions []JamSession

func (jamSession *JamSession) StartNextSong() {
	log.WithField("JamSession", jamSession.Label).Trace("Model event: Start next song for jamSession")

	song, err := jamSession.Queue.GetNextSong(true)

	if err != nil {
		log.WithField("JamSession", jamSession.Label).Trace("Model event: Queue is empty")
		return
	}

	jamSession.CurrentSong = &song.Song

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{jamSession.CurrentSong.URI},
	}

	err = jamSession.Client.PlayOpt(&playOptions)
	if err != nil {
		log.WithField("JamSession", jamSession.Label).Error("Error starting next song: ", err.Error())
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

	jamSession.Active = state //Old Backend has set the value with a delay of 2.5 seconds
	jamSession.PlaybackState.Playing = state
}

func (jamSession *JamSession) SetClientID(id spotify.ID) {
	if jamSession.DeviceID != id {
		playOptions := spotify.PlayOptions{DeviceID: &id}
		err := jamSession.Client.PlayOpt(&playOptions)

		if err != nil {

		} else {
			jamSession.DeviceID = id
		}
	}
}

func (jamSession *JamSession) SetJamSessionName() {
	jamSession.User.DisplayName = strings.Join([]string{jamSession.User.DisplayName, "'s Jam Session"}, "")
}
