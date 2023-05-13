package song

import (
	"time"

	"github.com/zmb3/spotify/v2"
)

type Song struct {
	ID    string
	Track *spotify.FullTrack
	Votes map[string]bool
	Date  time.Time
}

func New(t *spotify.FullTrack) *Song {
	return &Song{
		Track: t,
		ID:    string(t.ID),
		Votes: make(map[string]bool),
		Date:  time.Now(),
	}
}

func (s *Song) GetVotes() []string {
	var votes []string
	i := 0
	for v, voted := range s.Votes {
		if voted {
			votes = append(votes, v)
			i++
		}
	}
	return votes
}

func (s *Song) Vote(voteID string) bool {
	if _, exists := s.Votes[voteID]; !exists {
		// create new vote
		s.Votes[voteID] = true
	} else {
		// flip vote state
		s.Votes[voteID] = !s.Votes[voteID]
	}
	// return new vote state
	return s.Votes[voteID]
}

func (s *Song) HasVote(voteID string) bool {
	x, ok := s.Votes[voteID]
	return ok && x
}
