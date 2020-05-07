package models

import (
	"github.com/googollee/go-socket.io"
	"github.com/zmb3/spotify"
	"strings"
)

type Party struct {
	partyLabel string
	spotifyClient spotify.Client
	queue Queue
	socket *socketio.Server
	currentSong *spotify.FullTrack
	queueActive bool
	selectedDeviceId spotify.ID
	ipVoting bool
	playbackState spotify.PlayerState
	user spotify.User
}

type PartySettings struct {
	DeviceId spotify.ID `json:"device"`
	IpVoting bool `json:"ip"`
}

func (party *Party) setUser (user spotify.User) {
	party.user = user
	party.user.DisplayName = strings.Join([]string{party.user.DisplayName, "'s Jam Session"}, "")
}

func (party *Party) getUser () spotify.User {
	return party.user
}

func (party *Party) getLabel () string {
	return party.partyLabel
}

func (party *Party) getSpotifyClient () *spotify.Client {
	return &party.spotifyClient
}

func (party *Party) getSelectedDeviceId () spotify.ID {
	return party.selectedDeviceId
}

func (party *Party) getQueue () *Queue {
	return &party.queue
}

func (party *Party) setSetting (setting PartySettings) {
	if party.selectedDeviceId != setting.DeviceId {
		playOptions := spotify.PlayOptions{
			DeviceID:        &setting.DeviceId,
		}
		err := party.spotifyClient.PlayOpt(&playOptions)
		if err != nil {

		} else {
			party.selectedDeviceId = setting.DeviceId
		}
	}
	party.ipVoting = setting.IpVoting
}

func (party *Party) getCurrentSong() *spotify.FullTrack{
	return party.currentSong
}

func (party *Party) getPlaybackState() *spotify.PlayerState{
	return &party.playbackState
}

func (party *Party) setPlaybackState(playbackState spotify.PlayerState) {
	party.playbackState = playbackState
}

func (party *Party) getQueueActive() bool{
	return party.queueActive
}

func (party *Party) setQueueActive(state bool) {
	if state {
		party.spotifyClient.Play()
	} else {
		party.spotifyClient.Pause()
	}

	party.queueActive = state //Old Backend has set the value with a delay of 2.5 seconds
	party.getPlaybackState().Playing = state
	res := make(map[string]interface{})
	res["currentSong"] = party.getCurrentSong()
	res["state"] = party.getPlaybackState()
	party.socket.BroadcastToRoom("sessions", party.partyLabel, "playback", res)

}

func (party *Party) startNextSong() {
	party.currentSong = &party.queue.GetNextSong(true).Song //Will fail if queue is empty

	playOptions := spotify.PlayOptions{
		URIs:            []spotify.URI{party.getCurrentSong().URI},
	}

	party.spotifyClient.PlayOpt(&playOptions)

}


