package models

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/zmb3/spotify"
	"math/rand"
	"strings"
)

type PartyController struct {
	Partys []Party
	Count int32
	Socket socketio.Server
}

func (partyController *PartyController) generateNewParty (client spotify.Client, user spotify.User) string {

	party := Party{
		partyLabel:       "",
		spotifyClient:    client,
		queue:            Queue{},
		socket:           &partyController.Socket,
		currentSong:      nil,
		queueActive:      true,
		selectedDeviceId: "",
		ipVoting:         false,
		playbackState:    nil,
		user:             nil,
	}

	party.partyLabel = partyController.generateRandomLabel()
	party.setUser(user)
	partyController.Partys = append(partyController.Partys, party)

	return party.partyLabel
}

func (partyController *PartyController) generateRandomLabel() string {
	labelArr := make([]string, 5)
	possibleChars := "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

	for i := 0; i < 5; i++ {
		labelArr[i] = string(possibleChars[rand.Intn(len(possibleChars))])
	}

	label := strings.Join(labelArr, "")

	exits := false
	for _, party := range partyController.Partys {
		if party.partyLabel == label{
			exits = true
		}
	}

	if exits {
		return partyController.generateRandomLabel()
	} else {
		return label
	}
}

func (partyController *PartyController) getParty(label string) *Party {
	for _, party := range partyController.Partys {
		if party.partyLabel == label{
			return &party
		}
	}
	return nil
}

func (partyController *PartyController) setSocket(socket socketio.Server) {
	partyController.Socket = socket
}

func (partyController *PartyController) queueWorker () {
	for _, party := range partyController.Partys {

		state, _ := party.spotifyClient.PlayerState()
		party.setPlaybackState(state)

		if party.queueActive {
			if state.Progress > state.Item.Duration - 1000 || !state.Playing {
				party.startNextSong()
				party.socket.BroadcastToRoom("session", party.partyLabel, "queue", party.getQueue().GetObjectWithoutId(""))
				res := make(map[string]interface{})
				res["currentSong"] = party.getCurrentSong()
				res["state"] = party.getPlaybackState()
				party.socket.BroadcastToRoom("session", party.partyLabel, "playback", res)
			}
		}
	}
}