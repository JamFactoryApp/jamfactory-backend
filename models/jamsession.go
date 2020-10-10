package models

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/notifications"
	"github.com/jamfactoryapp/jamfactory-backend/types"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"sync"
	"time"
)

const JamLabelChars = "ABCDEFGHJKLMNOPQRSTUVWXYZ123456789"

type JamSession struct {
	Label         string
	Name          string
	VotingType    types.VotingType
	Active        bool
	Context       context.Context
	Queue         *Queue
	Client        spotify.Client
	DeviceID      spotify.ID
	CurrentSong   *spotify.FullTrack
	PlaybackState *spotify.PlayerState
	Room          *notifications.Room
	mux           sync.Mutex
}

type JamSessions []*JamSession

func NewJamSession(client spotify.Client, label string) (*JamSession, error) {
	user, err := client.CurrentUser()
	if err != nil {
		return nil, err
	}

	playback, err := client.PlayerState()
	if err != nil {
		return nil, err
	}

	jamSession := &JamSession{
		Label:         label,
		Name:          fmt.Sprintf("%s's JamSession", user.DisplayName),
		Queue:         &Queue{},
		VotingType:    types.SessionVotingType,
		Client:        client,
		Context:       context.Background(),
		DeviceID:      playback.Device.ID,
		PlaybackState: playback,
		Active:        playback.Device.ID != "",
		Room:          notifications.NewRoom(),
	}

	go jamSession.Room.OpenDoors()
	return jamSession, nil
}

func (jamSession *JamSession) Conductor() {
	for {
		select {
		case <-jamSession.Context.Done():
			log.Debug("Conductor leaves the jam session")
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

					message := types.GetQueueResponse{
						Queue: jamSession.Queue.GetObjectWithoutId(""),
					}
					jamSession.NotifyClients(&notifications.Message{
						Event:   notifications.Queue,
						Message: message,
					})
				}
			}

			message := types.SocketPlaybackState{
				Playback: *jamSession.PlaybackState,
			}
			jamSession.NotifyClients(&notifications.Message{
				Event:   notifications.Playback,
				Message: message,
			})
		}
	}
}

func (jamSession *JamSession) StartNextSong() {
	log.WithField("JamSession", jamSession.Label).Trace("Model event: Start next song for jamSession")

	devices, err := jamSession.Client.PlayerDevices()
	if err != nil {
		log.WithField("JamSession", jamSession.Label).Error("Error starting next song: ", err.Error())
	}

	activeDeviceAvailable := false
	for _, d := range devices {
		if d.Active {
			activeDeviceAvailable = true
			break
		}
	}

	if !activeDeviceAvailable {
		log.WithField("JamSession", jamSession.Label).Debug("No active device found, setting inactive")
		jamSession.SetClientID("")
		jamSession.SetJamSessionState(false)
		return
	}

	song, err := jamSession.Queue.GetNextSong()

	if err != nil {
		log.WithField("JamSession", jamSession.Label).Trace("Model event: Queue is empty")
		return
	}

	playOptions := spotify.PlayOptions{
		URIs: []spotify.URI{song.Song.URI},
	}

	err = jamSession.Client.PlayOpt(&playOptions)
	if err != nil {
		log.WithField("JamSession", jamSession.Label).Error("Error starting next song: ", err.Error())
		return
	}

	jamSession.CurrentSong = song.Song
	err = jamSession.Queue.AdvanceQueue()

	if err != nil {
		log.WithField("JamSession", jamSession.Label).Trace("Model event: Could not advance queue")
		return
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

	jamSession.Active = state
	jamSession.PlaybackState.Playing = state
}

func (jamSession *JamSession) SetClientID(id spotify.ID) {
	if jamSession.DeviceID != id {
		err := jamSession.Client.TransferPlayback(id, jamSession.Active)

		if err != nil {
			log.Debug("Error setting Device ID: ", err.Error())
		} else {
			jamSession.DeviceID = id
		}
	}
}

func (jamSession *JamSession) UpdatePlaybackState(state *spotify.PlayerState) {
	jamSession.Lock()
	defer jamSession.Unlock()
	jamSession.PlaybackState = state
}

func (jamSession *JamSession) UpdateCurrentSong(item *spotify.FullTrack) {
	jamSession.Lock()
	defer jamSession.Unlock()
	jamSession.CurrentSong = item
}

func (jamSession *JamSession) Lock() {
	jamSession.mux.Lock()
}

func (jamSession *JamSession) Unlock() {
	jamSession.mux.Unlock()
}

func (jamSession *JamSession) IntroduceClient(conn *websocket.Conn) {
	client := notifications.NewClient(jamSession.Room, conn)
	client.Room.Register <- client

	go client.Write()
	go client.Read()
}

func (jamSession *JamSession) NotifyClients(msg *notifications.Message) {
	if len(jamSession.Room.Clients) > 0 {
		jamSession.Room.Broadcast <- msg
	}
}
