package queue

import (
	"sort"

	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/song"
	"github.com/pkg/errors"
)

var (
	ErrQueueEmpty   = errors.New("queue is empty")
	ErrSongNotFound = errors.New("song not found")
)

const (
	HostVoteIdentifier string = "Host"
)

// Queue holds an ordered list of songs
type Queue interface {
	sort.Interface
	// Tracks returns the ordered list of songs
	Tracks() []types.Song
	// For returns the ordered list of songs from a specific user's perspective
	For(voteID string) []types.Song
	// Advance removes the first song in this Queue
	Advance() error
	// Advance returns the first song in this Queue
	GetNext() (song.Song, error)
	// GetHistory returns the history of the Queue
	GetHistory(voteID string) []types.Song
	// Vote toggles a vote on a song in this Queue
	Vote(songID string, voteID string, song interface{}) error
	// Delete deletes a song from this Queue
	Delete(songID string) error
}
