package song

import (
	"sync"
	"time"

	"github.com/zmb3/spotify"
)

type Song struct {
	sync.Mutex
	id    string
	track *spotify.FullTrack
	votes map[string]bool
	date  time.Time
}

func New(t *spotify.FullTrack) *Song {
	return &Song{
		track: t,
		id:    string(t.ID),
		votes: make(map[string]bool),
		date:  time.Now(),
	}
}

func (s *Song) ID() string {
	return s.id
}

func (s *Song) Song() *spotify.FullTrack {
	return s.track
}

func (s *Song) Votes() []string {
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

func (s *Song) Date() time.Time {
	return s.date
}

func (s *Song) SetDate(t time.Time) {
	s.Lock()
	defer s.Unlock()
	s.date = t
}

func (s *Song) Vote(voteID string) bool {
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

func (s *Song) HasVote(voteID string) bool {
	x, ok := s.votes[voteID]
	return ok && x
}
