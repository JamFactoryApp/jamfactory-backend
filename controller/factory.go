package controller

import (
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"math/rand"
	"strings"
	"time"
)

var parties models.Parties

func initFactory() {
	parties = make(models.Parties, 0)
	go QueueWorker()
}

func GenerateNewParty(client spotify.Client) (string, error) {
	queue := models.PartyQueue{}
	user, err := client.CurrentUser()
	playback, err := client.PlayerState()

	if err != nil {
		return "", err
	}

	party := models.Party{
		Label:         "",
		Queue:         &queue,
		IpVoteEnabled: false,
		Client:        client,
		DeviceID:      playback.Device.ID,
		CurrentSong:   playback.Item,
		PlaybackState: playback,
		User:          user,
		Active:        true,
	}

	party.Label = GenerateRandomLabel()
	party.SetPartyName()
	parties = append(parties, party)

	return party.Label, nil
}

func GenerateRandomLabel() string {
	labelSlice := make([]string, 5)
	possibleChars := "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

	for i := 0; i < 5; i++ {
		labelSlice[i] = string(possibleChars[rand.Intn(len(possibleChars))])
	}

	label := strings.Join(labelSlice, "")

	exits := false
	for _, party := range parties {
		if party.Label == label {
			exits = true
			break
		}
	}

	if exits {
		return GenerateRandomLabel()
	}

	return label
}

func GetParty(label string) *models.Party {
	for i := range parties {
		if parties[i].Label == label {
			return &parties[i]
		}
	}
	return nil
}

func QueueWorker() {
	for {
		time.Sleep(time.Second)

		for i := range parties {
			state, err := parties[i].Client.PlayerState()

			if err != nil {
				log.Printf("Couldn't get state for %s", parties[i].Label)
				continue
			}

			parties[i].PlaybackState = state
			parties[i].CurrentSong = state.Item

			if parties[i].Active {
				if !state.Playing || state.Progress > state.Item.Duration-1000 {
					log.Printf("Start next song for %s", parties[i].Label)
					parties[i].StartNextSong()
					Socket.BroadcastToRoom("/", parties[i].Label, SocketEventQueue, parties[i].Queue.GetObjectWithoutId(""))

					res := playbackBody{
						CurrentSong: parties[i].CurrentSong,
						Playback:    parties[i].PlaybackState,
					}

					Socket.BroadcastToRoom("/", parties[i].Label, SocketEventPlayback, res)
				}
			}
		}
	}
}
