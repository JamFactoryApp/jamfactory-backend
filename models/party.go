package models

import (
	"github.com/googollee/go-socket.io"
	"github.com/zmb3/spotify"
	"log"
	"strings"
)

type Party struct {
	Label         string
	Queue         *PartyQueue
	IpVoteEnabled bool
	Client        spotify.Client
	DeviceID      spotify.ID
	CurrentSong   *spotify.FullTrack
	PlaybackState spotify.PlayerState
	User          *spotify.PrivateUser
	Socket        *socketio.Server
}

type PartySettings struct {
	DeviceId spotify.ID `json:"device"`
	IpVoting bool       `json:"ip"`
}

func (party *Party) StartNextSong() {
	song, err := party.Queue.GetNextSong(true)

	if err != nil {
		return
	}

	party.CurrentSong = &song.Song //Will fail if queue is empty

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{party.CurrentSong.URI},
	}

	err = party.Client.PlayOpt(&playOptions)
	if err != nil {
		log.Println("Error starting next song")
	}
}

func (party *Party) SetUser(user *spotify.PrivateUser) {
	party.User = user
	party.User.DisplayName = strings.Join([]string{party.User.DisplayName, "'s Jam Session"}, "")
}

func (party *Party) SetSetting(setting PartySettings) {
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
}

func (party *Party) SetQueueActive(state bool) {
	if state {
		err := party.Client.Play()
		if err != nil {
			log.Println("Error setting client to play")
		}
	} else {
		err := party.Client.Pause()
		if err != nil {
			log.Println("Error setting client to pause")
		}
	}

	party.Queue.Active = state //Old Backend has set the value with a delay of 2.5 seconds
	party.PlaybackState.Playing = state
	res := make(map[string]interface{})
	res["currentSong"] = party.CurrentSong
	res["state"] = party.PlaybackState
	party.Socket.BroadcastToRoom("sessions", party.Label, "playback", res)

}
