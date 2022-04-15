package jamsession

import (
	"errors"
	"fmt"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/permissions"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/queue"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
)

const (
	UpdateIntervalInactive int = 10
	UpdateIntervalPlaying  int = 5
	UpdateIntervalSync     int = 1
)

var (
	ErrCollectionTypeInvalid     = errors.New("invalid collection type")
	ErrCouldNotGetAlbum          = errors.New("could not get album")
	ErrCouldNotGetAlbumTracks    = errors.New("could not get album tracks")
	ErrCouldNotGetPlaylistTracks = errors.New("could not get playlist tracks")
)

type JamSession struct {
	sync.Mutex
	users         *users.Store
	jamLabel      string
	name          string
	active        bool
	password      string
	members       Members
	lastTimestamp time.Time
	queue         *queue.Queue
	room          *notifications.Room
	quit          chan bool
}

func New(host *users.User, users *users.Store, label string) (*JamSession, error) {
	members := Members{
		host.Identifier: NewMember(host.Identifier, permissions.Guest, permissions.Host),
	}

	s := &JamSession{
		users:         users,
		jamLabel:      label,
		name:          fmt.Sprintf("%s's JamSession", host.UserName),
		active:        false,
		password:      "",
		members:       members,
		lastTimestamp: time.Now(),
		queue:         queue.New(),
		room:          notifications.NewRoom(),
		quit:          make(chan bool),
	}
	go s.room.OpenDoors()
	log.Info("Created new JamSession for ", host.UserName)
	return s, nil
}

func (s *JamSession) Conductor() {
	ticker := time.NewTicker(time.Second)
	syncCount := 0
	intervalCount := 0
	updateInterval := UpdateIntervalInactive
	defer ticker.Stop()
	for {
		select {

		// Fire conductor if he isn't needed anymore
		case <-s.quit:
			return

		// Update player state and send it to all connected clients
		case <-ticker.C:
			host, err := s.members.Host().ToUser(s.users)
			if err != nil {
				continue
			}
			intervalCount++
			if intervalCount >= updateInterval {
				intervalCount = 0
				playerState, err := host.Client().PlayerState()
				if err != nil {
					continue
				}
				host.SetPlayerState(playerState)
				s.SocketPlaybackUpdate(host)
				if !host.Synchronized {
					syncCount++
					if syncCount >= 3 {
						host.Synchronized = true
						syncCount = 0
					}
				}
			}
			// Check if the user started a song
			if s.synchronized && s.player.Item != nil && s.currentSong != nil && s.player.Item.ID != s.currentSong.ID {
				s.SetActive(false)
				s.currentSong = nil
				s.SocketJamUpdate()
			}

			// Check if no start or end of song is near
			if s.active {
				so, err := s.queue.GetNext()
				switch err {
				case nil:
					if (!s.player.Playing && s.player.Progress == 0) || (s.player.Item != nil && s.player.Progress > s.player.Item.Duration-1000) {
						if err := s.Play(s.player.Device, so.Song(), true); err != nil {
							log.Error(err)
							continue
						}
						s.SetTimestamp(time.Now())
					}
				case queue.ErrQueueEmpty:

				default:
					log.Error(err)
					break

				}
			}
			ticker.Reset(time.Second)
		}

		// Set the current update Interval
		if s.player.Playing && s.player.Item != nil {
			if s.player.Progress > s.player.Item.Duration-6000 {
				// First and last 5 seconds of the current song. Sync fast to correctly display switching the song
				updateInterval = UpdateIntervalSync
			} else {
				// We are in the middle of the song. Decrease sync rate
				updateInterval = UpdateIntervalPlaying
			}
		} else {
			// JamSession is inactive and no playback needs to be updated
			updateInterval = UpdateIntervalInactive
		}
		if !s.synchronized {
			// Conductor is not synchronized.
			updateInterval = UpdateIntervalSync
		}
	}
}
func (s *JamSession) Play(track *spotify.FullTrack, remove bool) error {
	host, err := s.members.Host().ToUser(s.users)
	if err != nil {
		return err
	}
	err = host.Play(track)
	if err != nil {
		return err
	}
	if remove {
		if err := s.queue.Delete(track.ID.String()); err != nil {
			return err
		}
	}

	s.SocketQueueUpdate()

	return nil
}

func (s *JamSession) JamLabel() string {
	return s.jamLabel
}

func (s *JamSession) Members() Members {
	return s.members
}

func (s *JamSession) Name() string {
	return s.name
}

func (s *JamSession) Active() bool {
	return s.active
}

func (s *JamSession) Password() string {
	return s.password
}

func (s *JamSession) SetPassword(password string) {
	s.Lock()
	defer s.Unlock()
	s.password = password
}

func (s *JamSession) Timestamp() time.Time {
	return s.lastTimestamp
}

func (s *JamSession) SetName(name string) {
	s.Lock()
	defer s.Unlock()
	s.name = name
}

func (s *JamSession) SetActive(active bool) {
	s.Lock()
	defer s.Unlock()
	s.active = active
}

func (s *JamSession) SetTimestamp(time time.Time) {
	s.Lock()
	defer s.Unlock()
	s.lastTimestamp = time
}

func (s *JamSession) Deconstruct() error {
	s.SetActive(false)
	s.room.CloseDoors()
	s.quit <- true
	return nil
}

func (s *JamSession) NotifyClients(msg *notifications.Message) {
	if len(s.room.Clients) > 0 {
		s.room.Broadcast <- msg
	}
}

func (s *JamSession) Queue() *queue.Queue {
	return s.queue
}

func (s *JamSession) AddCollection(collectionType string, collectionID string) error {
	host, err := s.members.Host().ToUser(s.users)
	if err != nil {
		return err
	}
	switch collectionType {
	case "playlist":
		playlist, err := host.Client().GetPlaylistTracks(spotify.ID(collectionID))

		if err != nil {
			return ErrCouldNotGetPlaylistTracks
		}

		for i := 0; i < len(playlist.Tracks); i++ {
			track, err := host.GetTrack(string(playlist.Tracks[i].Track.ID))
			if err != nil {
				return err
			}
			if err := s.queue.Vote(string(playlist.Tracks[i].Track.ID), queue.HostVoteIdentifier, track); err != nil {
				return err
			}
		}

	case "album":
		album, err := host.Client().GetAlbumTracks(spotify.ID(collectionID))

		if err != nil {
			return ErrCouldNotGetAlbum
		}

		ids := make([]spotify.ID, len(album.Tracks))
		for i := 0; i < len(album.Tracks); i++ {
			ids[i] = album.Tracks[i].ID
		}

		tracks, err := host.Client().GetTracks(ids...)
		if err != nil {
			return ErrCouldNotGetAlbumTracks
		}

		for i := 0; i < len(tracks); i++ {
			track, err := host.GetTrack(string(tracks[i].ID))
			if err != nil {
				return err
			}
			if err := s.queue.Vote(string(tracks[i].ID), queue.HostVoteIdentifier, track); err != nil {
				return err
			}
		}

	default:
		return ErrCollectionTypeInvalid
	}
	s.SocketQueueUpdate()
	return nil
}

func (s *JamSession) Vote(songID string, voteID string) error {
	host, err := s.members.Host().ToUser(s.users)
	if err != nil {
		return err
	}
	track, err := host.GetTrack(songID)
	if err != nil {
		return err
	}

	if err := s.queue.Vote(string(track.ID), voteID, track); err != nil {
		return err
	}
	s.SocketQueueUpdate()
	return nil
}

func (s *JamSession) Search(index string, searchType spotify.SearchType, options *spotify.Options) (interface{}, error) {
	host, err := s.members.Host().ToUser(s.users)
	if err != nil {
		return nil, err
	}
	return host.Search(index, searchType, options)
}

func (s *JamSession) IntroduceClient(conn *websocket.Conn) {
	client := notifications.NewClient(s.room, conn)
	client.Room.Register <- client

	go client.Write()
	go client.Read()
}

func (s *JamSession) DeleteSong(songID string) error {
	if err := s.queue.Delete(songID); err != nil {
		return err
	}
	s.SocketQueueUpdate()
	return nil
}

func (s *JamSession) SocketJamUpdate() {
	s.NotifyClients(&notifications.Message{
		Event: notifications.Jam,
		Message: types.SocketJamMessage{
			Label:  s.jamLabel,
			Name:   s.name,
			Active: s.active,
		},
	})
}

func (s *JamSession) SocketQueueUpdate() {
	s.NotifyClients(&notifications.Message{
		Event: notifications.Queue,
		Message: types.PutQueuePlaylistsResponse{
			Tracks: s.Queue().Tracks(),
		},
	})
}

func (s *JamSession) SocketPlaybackUpdate(host *users.User) {
	s.NotifyClients(&notifications.Message{
		Event: notifications.Playback,
		Message: types.SocketPlaybackMessage{
			Playback: host.GetPlayerState(),
			DeviceID: host.GetPlayerState().Device.ID,
		},
	})
}
