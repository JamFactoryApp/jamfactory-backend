package controllers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
	"jamfactory-backend/types"
	"math/rand"
	"strings"
	"time"
)

var jamSessions models.JamSessions

func initFactory() {
	jamSessions = make(models.JamSessions, 0)
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
		VotingType:    types.SessionVotingType,
		Client:        client,
		Context:       context.Background(),
		DeviceID:      playback.Device.ID,
		CurrentSong:   nil,
		PlaybackState: playback,
		Active:        true,
	}

	if jamSession.DeviceID == "" {
		jamSession.Active = false
	}

	jamSessions = append(jamSessions, &jamSession)

	go Conductor(&jamSession)
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

func Conductor(jamSession *models.JamSession) {
	for {
		select {
		case <-jamSession.Context.Done():
			log.Debug("Conductors leaves the jam session")
			return
		default:
			time.Sleep(time.Second)
			state, err := jamSession.Client.PlayerState()

			if err != nil {
				log.WithField("Label", jamSession.Label).Debug("Conductor couldn't get state for")
				continue
			}

			jamSession.UpdatePlaybackState(state)
			jamSession.UpdateCurrentSong(state.Item)

			if jamSession.DeviceID == "" && state.Device != (spotify.PlayerDevice{}) {
				jamSession.DeviceID = state.Device.ID
			}

			if jamSession.Active && jamSession.Queue.Len() > 0 && jamSession.DeviceID != "" {
				if !state.Playing || state.Progress > state.Item.Duration-1000 {
					log.WithField("Label", jamSession.Label).Debug("Conductor started next song for")
					jamSession.StartNextSong()
					SendToRoom(jamSession.Label, SocketEventQueue, jamSession.Queue.GetObjectWithoutId(""))
				}
			}

			res := types.SocketPlaybackState{
				Playback: *jamSession.PlaybackState,
			}
			SendToRoom(jamSession.Label, SocketEventPlayback, res)
		}
	}
}
