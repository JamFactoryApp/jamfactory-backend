package models

import (
	"github.com/zmb3/spotify"
	"time"
)

type Song struct {
	Song spotify.FullTrack
	Votes []Vote
	Date time.Time
}

type SongWithoutId struct {
	Song spotify.FullTrack
	Votes int
	Voted bool
}

func (song *Song) GetVoteCount() int{
	return len(song.Votes)
}

func (song *Song) Vote (id string) bool {
	if containsVote(song.Votes, id) {
		// The id has already voted for the song. Another vote will remove the current vote
		i := getIndexOfVote(song.Votes, id)
		song.Votes = append(song.Votes[:i], song.Votes[i+1:]...)
		return false
	}

	// The id has currently not voted for the song. Add the vote
	vote := Vote{id: id}
	song.Votes = append(song.Votes, vote)
	return true
}

func (song *Song) CheckForVote (id string) bool {
	if getIndexOfVote(song.Votes, id) != -1 {
		return true
	}
	return false
}

func (song *Song) GetObjectWithoutId(id string) SongWithoutId {
	return SongWithoutId{
		Song:  song.Song,
		Votes: song.GetVoteCount(),
		Voted: song.CheckForVote(id),
	}
}

func containsVote(arr []Vote, id string) bool {
	for _, a := range arr {
		if a.id == id {
			return true
		}
	}
	return false
}

func getIndexOfVote(arr []Vote, id string) int {
	for i, a := range arr {
		if a.id == id {
			return i
		}
	}
	return -1
}