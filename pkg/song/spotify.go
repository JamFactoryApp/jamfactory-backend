package song

import (
	"github.com/zmb3/spotify"
	"sync"
	"time"
)

type SpotifySong struct {
	sync.Mutex
	id    string
	track *spotify.FullTrack
	votes map[string]bool
	date  time.Time
}

func NewSpotify(t *spotify.FullTrack) Song {
	return &SpotifySong{
		track: t,
		id:    string(t.ID),
		votes: make(map[string]bool),
		date:  time.Now(),
	}
}

func (s *SpotifySong) ID() string {
	return s.id
}

func (s *SpotifySong) Song() interface{} {
	return s.track
}

func (s *SpotifySong) Votes() []string {
	var votes []string
	i := 0
	for v, voted := range s.votes {
		if voted {
			votes = append(votes, v)
			i++
		}
	}
	return votes
}

func (s *SpotifySong) Date() time.Time {
	return s.date
}

func (s *SpotifySong) SetDate(t time.Time) {
	s.Lock()
	defer s.Unlock()
	s.date = t
}

func (s *SpotifySong) Vote(voteID string) bool {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.votes[voteID]; !exists {
		// create new vote
		s.votes[voteID] = true
	} else {
		// flip vote state
		s.votes[voteID] = !s.votes[voteID]
	}

	// return new vote state
	return s.votes[voteID]
}

func (s *SpotifySong) HasVote(voteID string) bool {
	x, ok := s.votes[voteID]
	return ok && x
}
