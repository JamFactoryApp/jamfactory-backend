package controllers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"math/rand"
	"strings"
	"time"
)

var jamSessions models.JamSessions

func initFactory() {
	jamSessions = make(models.JamSessions, 0)
	go QueueWorker()
}

func GenerateNewJamSession(client spotify.Client) (string, error) {
	queue := models.Queue{}
	user, err := client.CurrentUser()
	playback, err := client.PlayerState()

	if err != nil {
		return "", err
	}

	jamSession := models.JamSession{
		Label:         GenerateRandomLabel(),
		Name:          strings.Join([]string{user.DisplayName, "'s Jam Session"}, ""),
		Queue:         &queue,
		IpVoteEnabled: false,
		Client:        client,
		Context:       context.Background(),
		DeviceID:      playback.Device.ID,
		CurrentSong:   nil,
		PlaybackState: playback,
		Active:        true,
	}
	jamSessions = append(jamSessions, &jamSession)

	return jamSession.Label, nil
}

func GenerateRandomLabel() string {
	labelSlice := make([]string, 5)
	possibleChars := "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

	for i := 0; i < 5; i++ {
		labelSlice[i] = string(possibleChars[rand.Intn(len(possibleChars))])
	}

	label := strings.Join(labelSlice, "")

	exists := false
	for _, jamSession := range jamSessions {
		if jamSession.Label == label {
			exists = true
			break
		}
	}

	if exists {
		return GenerateRandomLabel()
	}

	return label
}

func GetJamSession(label string) *models.JamSession {
	for i := range jamSessions {
		if jamSessions[i].Label == label {
			return jamSessions[i]
		}
	}
	return nil
}

func DeleteJamSession(label string) {
	for i := range jamSessions {
		if jamSessions[i].Label == label {
			jamSessions[i].SetJamSessionState(false)
			ctx, cancel := context.WithCancel(jamSessions[i].Context)
			jamSessions[i].Context = ctx
			jamSessions = append(jamSessions[:i], jamSessions[i+1:]...)
			cancel()
		}
	}
}

func QueueWorker() {
	for {
		time.Sleep(time.Second)

		for i := range jamSessions {
			state, err := jamSessions[i].Client.PlayerState()

			if err != nil {
				log.Printf("Couldn't get state for %s", jamSessions[i].Label)
				continue
			}

			jamSessions[i].PlaybackState = state
			jamSessions[i].CurrentSong = state.Item

			if jamSessions[i].Active && jamSessions[i].Queue.Len() > 0 {
				if !state.Playing || state.Progress > state.Item.Duration-1000 {
					log.Printf("Start next song for %s", jamSessions[i].Label)
					jamSessions[i].StartNextSong()
					Socket.BroadcastToRoom(SocketNamespace, jamSessions[i].Label, SocketEventQueue, jamSessions[i].Queue.GetObjectWithoutId(""))

					res := playbackBody{
						Playback:    jamSessions[i].PlaybackState,
					}

					Socket.BroadcastToRoom(SocketNamespace, jamSessions[i].Label, SocketEventPlayback, res)
				}
			}
		}
	}
}
