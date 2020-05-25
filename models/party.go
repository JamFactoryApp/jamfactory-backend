package models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"strings"
)

type Party struct {
	Label         string
	Queue         *PartyQueue
	IpVoteEnabled bool
	Client        spotify.Client
	DeviceID      spotify.ID
	CurrentSong   *spotify.FullTrack
	PlaybackState *spotify.PlayerState
	User          *spotify.PrivateUser
	Active        bool
}

type Parties []Party

func (party *Party) StartNextSong() {
	log.WithField("Party", party.Label).Trace("Model event: Start next song for party")

	song, err := party.Queue.GetNextSong(true)

	if err != nil {
		log.WithField("Party", party.Label).Trace("Model event: Queue is empty")
		return
	}

	party.CurrentSong = &song.Song

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{party.CurrentSong.URI},
	}

	err = party.Client.PlayOpt(&playOptions)
	if err != nil {
		log.WithField("Party", party.Label).Error("Error starting next song: ", err.Error())
	}
}

func (party *Party) SetPartyState(state bool) {
	log.WithField("Party", party.Label).Trace("Model event: Set party enabled")

	var fragment string
	var err error

	if state {
		err = party.Client.Play()
		fragment = "play"
	} else {
		err = party.Client.Pause()
		fragment = "pause"
	}

	if err != nil {
		log.WithField("Party", party.Label).Warn(fmt.Sprintf("Error setting client to %s\n", fragment))
	}

	party.Active = state //Old Backend has set the value with a delay of 2.5 seconds
	party.PlaybackState.Playing = state
}

func (party *Party) SetClientID(id spotify.ID) {
	if party.DeviceID != id {
		playOptions := spotify.PlayOptions{DeviceID: &id}
		err := party.Client.PlayOpt(&playOptions)

		if err != nil {

		} else {
			party.DeviceID = id
		}
	}
}

func (party *Party) SetPartyName() {
	party.User.DisplayName = strings.Join([]string{party.User.DisplayName, "'s Jam Session"}, "")
}
