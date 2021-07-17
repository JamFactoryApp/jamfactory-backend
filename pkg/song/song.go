package song

import (
	"time"
)

// Song holds metadata for a song that can be played in a JamSession
type Song interface {
	// ID returns a unique id for this song
	ID() string
	// Song returns an object containing data specific to the music streaming provider
	Song() interface{}
	// Votes returns all votes that have been cast on this song
	Votes() []string
	// Vote toggles a vote on this song
	Vote(voteID string) bool
	// Date returns the time this song was first voted on
	Date() time.Time
	// SetDate updates this Song's Date
	SetDate(t time.Time)
	// HasVote returns whether a given voteID has voted on this Song
	HasVote(voteID string) bool
}
