package models

import (
	"github.com/zmb3/spotify"
	"time"
)

type Song struct {
	Song  spotify.FullTrack
	Votes []Vote
	Date  time.Time
}

type SongWithoutId struct {
	Song  spotify.FullTrack
	Votes int
	Voted bool
}

func (song *Song) Vote(id string) bool {
	if song.containsVote(id) {
		// The id has already voted for the song. Another vote will remove the current vote
		i := song.indexOfVote(id)
		song.Votes = append(song.Votes[:i], song.Votes[i+1:]...)
		return false
	}

	// The id has currently not voted for the song. Add the vote
	vote := Vote{id: id}
	song.Votes = append(song.Votes, vote)
	return true
}

func (song Song) VoteCount() int {
	return len(song.Votes)
}

func (song Song) WithoutId(voteID string) SongWithoutId {
	return SongWithoutId{
		Song:  song.Song,
		Votes: song.VoteCount(),
		Voted: song.containsVote(voteID),
	}
}

func (song Song) containsVote(id string) bool {
	for _, a := range song.Votes {
		if a.id == id {
			return true
		}
	}
	return false
}

func (song Song) indexOfVote(id string) int {
	for i, a := range song.Votes {
		if a.id == id {
			return i
		}
	}
	return -1
}
