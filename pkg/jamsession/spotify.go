package jamsession

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jamfactoryapp/jamfactory-backend/pkg/hub"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/permissions"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/store"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/users"

	"github.com/gorilla/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/queue"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify/v2"
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

type Stores struct {
	Members  store.Store[Members]
	Queues   store.Store[queue.Queue]
	Settings store.Store[Settings]
}

type Settings struct {
	Name     string
	Active   bool
	Password string
}

type JamLabel string

type JamSession struct {
	JamLabel  string
	stores    Stores
	hub       *hub.Hub
	Timestamp time.Time
	room      *notifications.Room
	quit      chan bool
}

func CreateNew(host *users.User, stores Stores, hub *hub.Hub, label string) (*JamSession, error) {
	members := &Members{
		host.Identifier: NewMember(host.Identifier, permissions.Guest, permissions.Host),
	}
	userInfo, err := host.GetInfo()
	if err != nil {
		return nil, err
	}

	settings := &Settings{
		Name:     fmt.Sprintf("%s's JamSession", userInfo.UserName),
		Active:   false,
		Password: "",
	}

	currentQueue := queue.New()

	s := &JamSession{
		JamLabel:  label,
		Timestamp: time.Now(),
		hub:       hub,
		room:      notifications.NewRoom(),
		quit:      make(chan bool),
		stores:    stores,
	}

	if err = s.SetMembers(members); err != nil {
		return nil, err
	}
	if err = s.SetSettings(settings); err != nil {
		return nil, err
	}
	if err = s.SetQueue(currentQueue); err != nil {
		return nil, err
	}

	go s.room.OpenDoors()
	go s.Conductor()
	log.WithField("Label", label).Info("Created new JamSession")
	return s, nil
}

func Load(stores Stores, hub *hub.Hub, label string) (*JamSession, error) {
	s := &JamSession{
		JamLabel:  label,
		Timestamp: time.Now(),
		hub:       hub,
		room:      notifications.NewRoom(),
		quit:      make(chan bool),
		stores:    stores,
	}
	go s.Conductor()
	go s.room.OpenDoors()
	log.WithField("Label", label).Info("Loaded JamSession from store")
	return s, nil
}

func (s *JamSession) GetQueue() (*queue.Queue, error) {
	return s.stores.Queues.Get(s.JamLabel)
}

func (s *JamSession) SetQueue(queue *queue.Queue) error {
	return s.stores.Queues.Save(queue, s.JamLabel)
}

func (s *JamSession) GetMembers() (*Members, error) {
	return s.stores.Members.Get(s.JamLabel)
}

func (s *JamSession) SetMembers(members *Members) error {
	return s.stores.Members.Save(members, s.JamLabel)
}

func (s *JamSession) GetSettings() (*Settings, error) {
	return s.stores.Settings.Get(s.JamLabel)
}

func (s *JamSession) SetSettings(settings *Settings) error {
	return s.stores.Settings.Save(settings, s.JamLabel)
}

func (s *JamSession) Conductor() {
	ticker := time.NewTicker(time.Second)
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
			members, err := s.GetMembers()
			if err != nil {
				log.Warn(err)
			}
			settings, err := s.GetSettings()
			if err != nil {
				log.Warn(err)
			}
			currentQueue, err := s.GetQueue()
			if err != nil {
				log.Warn(err)
			}
			// Get the host user
			hostMember, err := members.Host()
			if err != nil {
				continue
			}
			host, err := s.hub.GetUserByIdentifier(context.Background(), hostMember.Identifier)
			if err != nil {
				continue
			}

			// Go to all members joined by the JamSession
			for _, member := range *members {
				// Get the user for the member
				user, err := s.hub.GetUserByIdentifier(context.Background(), member.Identifier)
				if err != nil {
					log.Warn(err)
					continue
				}
				userInfo, err := user.GetInfo()
				if err != nil {
					log.Warn(err)
					continue
				}
				// Conductor operation is only relevant for spotify users
				if userInfo.UserType != users.UserTypeSpotify {
					continue
				}
				// If the intervalCount is reached, update the PlayerState for each spotify user
				if intervalCount >= updateInterval {

					playerState, err := user.Client().PlayerState(context.Background())
					if err != nil {
						continue
					}

					user.SetPlayerState(playerState)

					if !user.Synchronized {
						user.SyncCount++
						if user.SyncCount >= 1 {
							user.Synchronized = true
							user.SyncCount = 0
						}
					}
				}

				// Check if the user started a song
				if user.Synchronized && user.GetPlayerState().Item != nil && user.CurrentSong != nil && user.GetPlayerState().Item.ID != user.CurrentSong.ID {
					user.Active = false
					user.CurrentSong = nil
					if user.Identifier == host.Identifier {
						settings.Active = false
						if err := s.SetSettings(settings); err != nil {
							log.Warn(err)
							continue
						}
						s.SocketJamUpdate()
					}
				}

			}

			s.SocketPlaybackUpdate(host)

			// Check if no start or end of song is near for the host
			if settings.Active && host.Synchronized {
				so, err := currentQueue.GetNext()
				switch err {
				case nil:
					if (!host.GetPlayerState().Playing && host.GetPlayerState().Progress == 0) ||
						(host.GetPlayerState().Item != nil && host.GetPlayerState().Progress > host.GetPlayerState().Item.Duration-1000) {

						for _, member := range *members {
							// Get the user for the member
							user, err := s.hub.GetUserByIdentifier(context.Background(), member.Identifier)
							if err != nil {
								log.Warn(err)
								continue
							}
							userInfo, err := user.GetInfo()
							if err != nil {
								log.Warn(err)
								continue
							}
							// Conductor operation is only relevant for spotify users
							if userInfo.UserType != users.UserTypeSpotify {
								continue
							}

							if member.HasPermissions(permissions.Host) || (member.HasPermissions(permissions.Listen) && userInfo.UserStartListening) {
								if err := user.Play(context.Background(), so.Track); err != nil {
									log.Error(err)
									continue
								}
							}
						}

						currentQueue.Delete(so.ID)
						err = s.SetQueue(currentQueue)
						s.Timestamp = time.Now()
					}
				case queue.ErrQueueEmpty:

				default:
					log.Error(err)
				}
			}

			// Reset the interval count
			if intervalCount >= updateInterval {
				intervalCount = 0
			} else {
				intervalCount++
			}

			// Set the current update Interval
			if host.GetPlayerState().Playing && host.GetPlayerState().Item != nil {
				if host.GetPlayerState().Progress > host.GetPlayerState().Item.Duration-6000 {
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
			if !host.Synchronized {
				// Conductor is not synchronized.
				updateInterval = UpdateIntervalSync
			}

			ticker.Reset(time.Second)
		}
	}
}
func (s *JamSession) Play(ctx context.Context, track *spotify.FullTrack, remove bool) error {
	members, err := s.GetMembers()
	currentQueue, err := s.GetQueue()
	if err != nil {
		return err
	}
	hostMember, err := members.Host()
	if err != nil {
		return err
	}
	host, err := s.hub.GetUserByIdentifier(ctx, hostMember.Identifier)
	if err != nil {
		return err
	}
	err = host.Play(ctx, track)
	if err != nil {
		return err
	}
	if remove {
		currentQueue.Delete(track.ID.String())
	}
	err = s.SetQueue(currentQueue)
	if err != nil {
		return err
	}
	s.SocketQueueUpdate()

	return nil
}

func (s *JamSession) Deconstruct() error {
	s.room.CloseDoors()
	s.quit <- true
	return nil
}

func (s *JamSession) NotifyClients(msg *notifications.Message) {
	if len(s.room.Clients) > 0 {
		s.room.Broadcast <- msg
	}
}

func (s *JamSession) AddCollection(ctx context.Context, collectionType string, collectionID string) error {
	members, err := s.GetMembers()
	currentQueue, err := s.GetQueue()
	if err != nil {
		return err
	}
	hostMember, err := members.Host()
	if err != nil {
		return err
	}
	host, err := s.hub.GetUserByIdentifier(ctx, hostMember.Identifier)
	if err != nil {
		return err
	}
	switch collectionType {
	case "playlist":
		playlist, err := host.Client().GetPlaylistItems(ctx, spotify.ID(collectionID))
		if err != nil {
			return ErrCouldNotGetPlaylistTracks
		}

		for _, item := range playlist.Items {
			if item.Track.Track == nil {
				log.Printf("%+v", item)
				continue
			}
			if err := currentQueue.Vote(string(item.Track.Track.ID), queue.HostVoteIdentifier, item.Track.Track); err != nil {
				return err
			}
		}

	case "album":
		album, err := host.Client().GetAlbumTracks(ctx, spotify.ID(collectionID))

		if err != nil {
			return ErrCouldNotGetAlbum
		}

		ids := make([]spotify.ID, len(album.Tracks))
		for i := 0; i < len(album.Tracks); i++ {
			ids[i] = album.Tracks[i].ID
		}

		tracks, err := host.Client().GetTracks(ctx, ids)
		if err != nil {
			return ErrCouldNotGetAlbumTracks
		}

		for i := 0; i < len(tracks); i++ {
			track, err := host.GetTrack(ctx, string(tracks[i].ID))
			if err != nil {
				return err
			}
			if err := currentQueue.Vote(string(tracks[i].ID), queue.HostVoteIdentifier, track); err != nil {
				return err
			}
		}

	default:
		return ErrCollectionTypeInvalid
	}
	err = s.SetQueue(currentQueue)
	if err != nil {
		return err
	}
	s.SocketQueueUpdate()
	return nil
}

func (s *JamSession) Vote(ctx context.Context, songID string, voteID string) error {
	members, err := s.GetMembers()
	currentQueue, err := s.GetQueue()
	if err != nil {
		return err
	}
	hostMember, err := members.Host()
	if err != nil {
		return err
	}
	host, err := s.hub.GetUserByIdentifier(ctx, hostMember.Identifier)
	if err != nil {
		return err
	}
	track, err := host.GetTrack(ctx, songID)
	if err != nil {
		return err
	}

	if err := currentQueue.Vote(string(track.ID), voteID, track); err != nil {
		return err
	}
	err = s.SetQueue(currentQueue)
	if err != nil {
		return err
	}
	s.SocketQueueUpdate()
	return nil
}

func (s *JamSession) Search(ctx context.Context, index string, searchType spotify.SearchType, options ...spotify.RequestOption) (interface{}, error) {
	members, err := s.GetMembers()
	if err != nil {
		return nil, err
	}
	hostMember, err := members.Host()
	if err != nil {
		return nil, err
	}
	host, err := s.hub.GetUserByIdentifier(ctx, hostMember.Identifier)
	if err != nil {
		return nil, err
	}
	return host.Search(ctx, index, searchType, options...)
}

func (s *JamSession) IntroduceClient(conn *websocket.Conn) {
	client := notifications.NewClient(s.room, conn)
	client.Room.Register <- client

	go client.Write()
	go client.Read()
}

func (s *JamSession) DeleteSong(songID string) error {
	queue, err := s.GetQueue()
	if err != nil {
		return err
	}
	queue.Delete(songID)
	if err := s.SetQueue(queue); err != nil {
		return err
	}
	s.SocketQueueUpdate()
	return nil
}

func (s *JamSession) SocketJamUpdate() {
	settings, err := s.GetSettings()
	if err != nil {
		log.Warn("could not get settings", err)
		return
	}
	s.NotifyClients(&notifications.Message{
		Event: notifications.Jam,
		Message: types.SocketJamMessage{
			Label:  s.JamLabel,
			Name:   settings.Name,
			Active: settings.Active,
		},
	})
}

func (s *JamSession) SocketQueueUpdate() {
	queue, err := s.GetQueue()
	if err != nil {
		log.Warn("could not get settings", err)
		return
	}
	s.NotifyClients(&notifications.Message{
		Event: notifications.Queue,
		Message: types.PutQueuePlaylistsResponse{
			Tracks: queue.Tracks(),
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
