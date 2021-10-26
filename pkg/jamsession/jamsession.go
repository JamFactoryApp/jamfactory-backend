package jamsession

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/notifications"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/queue"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/song"
	"github.com/pkg/errors"
	"github.com/zmb3/spotify"
)

var (
	ErrJamSessionMissing   = errors.New("no JamSession provided")
	ErrJamSessionMalformed = errors.New("malformed JamSession")
)

// TODO: abstract out spotify
// JamSession is a private party with one host to set it up, and many attendees to join in.
type JamSession interface {
	// Conductor controls a JamSession. This method should run in a seperate goroutine
	Conductor()
	// JamLabel returns this JamSession's JamLabel
	JamLabel() string
	// Name returns this JamSession's name
	Name() string
	// SetName updates this JamSession's name
	SetName(name string)
	// Members returns the JamSession's members
	Members() Members
	// Active returns whether this JamSession is active
	Active() bool
	// SetActive activates or deactivates this JamSession
	SetActive(active bool)
	// Password returns the current password for the JamSession
	Password() string
	// SetPassword sets the current password for the JamSession
	SetPassword(password string)
	// Timestamp returns the last timestamp set for the JamSession
	Timestamp() time.Time
	// SetTimestamp updates the timestamp of the JamSession
	SetTimestamp(time.Time)
	// SetState updates this JamSession's playback state
	SetState(state bool) error
	// Deconstruct this JamSession
	Deconstruct() error
	// NotifyClients notifies this JamSession's guests using websockets
	NotifyClients(msg *notifications.Message)
	// Queue returns this JamSession's queue
	Queue() queue.Queue
	// AddCollection adds a collection such as a playlist or an album to this JamSession's queue
	AddCollection(collectionType string, collectionID string) error
	// IntroduceClient adds a new guest to this JamSession's notifications room
	IntroduceClient(conn *websocket.Conn)
	// Vote lets a user vote for a song in this JamSession's queue
	Vote(songID string, voteID string) error
	// DeleteSong removes a song from this JamSession's queue
	DeleteSong(songID string) error
	// Play plays a song
	Play(device spotify.PlayerDevice, song song.Song) error
	// Search TODO
	Search(index string, searchType spotify.SearchType, options *spotify.Options) (interface{}, error)
	// Playlists TODO
	Playlists() (*spotify.SimplePlaylistPage, error)
	// CreatePlaylist TODO
	CreatePlaylist(name string, desc string, songs []spotify.ID) error
	// Devices TODO
	Devices() ([]spotify.PlayerDevice, error)
	// GetSong TODO
	GetSong(songID string) (song.Song, error)
	// CurrentSong TODO
	CurrentSong() *spotify.FullTrack
	// GetPlayerState TODO
	GetPlayerState() *spotify.PlayerState
	// SetPlayerState TODO
	SetPlayerState(*spotify.PlayerState)
	// GetDeviceID TODO
	GetDevice() spotify.PlayerDevice
	// SetDeviceID TODO
	SetDevice(string) error
	// SocketJamUpdate updates the JamSession through the Websocket connection
	SocketJamUpdate()
	// SocketQueueUpdate updates the Queue through the Websocket connection
	SocketQueueUpdate()
	// SocketPlaybackUpdate updates the Playback through the Websocket connection
	SocketPlaybackUpdate()
}

type contextKey string

const key contextKey = "JamSession"

// NewContext returns a new context containing a JamSession
func NewContext(ctx context.Context, jamSession JamSession) context.Context {
	return context.WithValue(ctx, key, jamSession)
}

// FromContext returns a JamSession existing in a context
func FromContext(ctx context.Context) (JamSession, error) {
	val := ctx.Value(key)
	if val == nil {
		return nil, ErrJamSessionMissing
	}
	jamSession, ok := val.(JamSession)
	if !ok {
		return nil, ErrJamSessionMalformed
	}
	return jamSession, nil
}
