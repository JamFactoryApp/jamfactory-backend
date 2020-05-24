package models

import (
	"github.com/googollee/go-socket.io"
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
	Socket        *socketio.Server
	Active        bool
}

type PartySettings struct {
	DeviceId  spotify.ID
	IpVoting  bool
	PartyName string
}

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

func (party *Party) SetUser(user *spotify.PrivateUser) {
	log.WithField("Party", party.Label).Trace("Model event: Set party user")

	party.User = user
	party.User.DisplayName = strings.Join([]string{party.User.DisplayName, "'s Jam Session"}, "")
}

func (party *Party) SetSetting(setting PartySettings) {
	log.WithField("Party", party.Label).Trace("Model event: Set party settings")

	if party.DeviceID != setting.DeviceId {
		playOptions := spotify.PlayOptions{
			DeviceID: &setting.DeviceId,
		}
		err := party.Client.PlayOpt(&playOptions)
		if err != nil {

		} else {
			party.DeviceID = setting.DeviceId
		}
	}
	party.IpVoteEnabled = setting.IpVoting
	party.User.DisplayName = setting.PartyName
}

func (party *Party) SetPartyState(state bool) {
	log.WithField("Party", party.Label).Trace("Model event: Set party state")

	if state {
		err := party.Client.Play()
		if err != nil {
			log.WithField("Party", party.Label).Warn("Error setting client to play")
		}
	} else {
		err := party.Client.Pause()
		if err != nil {
			log.WithField("Party", party.Label).Warn("Error setting client to pause")
		}
	}

	party.Active = state //Old Backend has set the value with a delay of 2.5 seconds
	party.PlaybackState.Playing = state
	res := make(map[string]interface{})
	res["currentSong"] = party.CurrentSong
	res["state"] = party.PlaybackState
	party.Socket.BroadcastToRoom("sessions", party.Label, "playback", res)

}
