package controller

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"log"
	"math/rand"
	"strings"
)

type PartyController struct {
	Partys []models.Party
	Count  int32
	Socket *socketio.Server
}

func (pc *PartyController) generateNewParty(client spotify.Client) (string, error) {

	queue := models.PartyQueue{Active: true}
	user, err := client.CurrentUser()
	playback, err := client.PlayerState()

	if err != nil {
		return "", err
	}

	party := models.Party{
		Label:         "",
		Client:        client,
		Queue:         &queue,
		Socket:        pc.Socket,
		CurrentSong:   playback.Item,
		DeviceID:      playback.Device.ID,
		IpVoteEnabled: false,
		PlaybackState: playback,
		User:          user,
	}


	party.Label = pc.GenerateRandomLabel()
	party.User = user
	pc.Partys = append(pc.Partys, party)

	return party.Label, nil
}

func (pc *PartyController) GenerateRandomLabel() string {
	labelArr := make([]string, 5)
	possibleChars := "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

	for i := 0; i < 5; i++ {
		labelArr[i] = string(possibleChars[rand.Intn(len(possibleChars))])
	}

	label := strings.Join(labelArr, "")

	exits := false
	for _, party := range pc.Partys {
		if party.Label == label {
			exits = true
		}
	}

	if exits {
		return pc.GenerateRandomLabel()
	} else {
		return label
	}
}

func (pc *PartyController) GetParty(label string) *models.Party {
	for _, party := range pc.Partys {
		if party.Label == label {
			return &party
		}
	}
	return nil
}

func (pc *PartyController) SetSocket(socket *socketio.Server) {
	pc.Socket = socket
}

func QueueWorker(controller *PartyController) {
	for i := 0; i< len(controller.Partys); i++ {

		state, err := controller.Partys[i].Client.PlayerState()

		if err != nil {
			log.Printf("Couldn't get state for %s", controller.Partys[i].Label)
			continue
		}

		controller.Partys[i].PlaybackState = state
		controller.Partys[i].CurrentSong = state.Item

		if controller.Partys[i].Queue.Active {

			if !state.Playing || state.Progress > state.Item.Duration-1000 {
				log.Printf("Start next song for %s", controller.Partys[i].Label)
				controller.Partys[i].StartNextSong()
				controller.Partys[i].Socket.BroadcastToRoom("/", controller.Partys[i].Label, "queue", controller.Partys[i].Queue.GetObjectWithoutId(""))
				res := make(map[string]interface{})
				res["currentSong"] = controller.Partys[i].CurrentSong
				res["state"] = controller.Partys[i].PlaybackState
				controller.Partys[i].Socket.BroadcastToRoom("/", controller.Partys[i].Label, "playback", res)
			}
		}
	}
}
