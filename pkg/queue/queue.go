package queue

import (
	"errors"
	"sort"
	"time"

	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/song"
	"github.com/zmb3/spotify"
)

var (
	ErrQueueEmpty   = errors.New("queue is empty")
	ErrSongNotFound = errors.New("song not found")
)

const (
	HostVoteIdentifier string = "Host"
)

type Queue struct {
	Songs   []*song.Song
	History []*song.Song
}

func New() *Queue {
	return &Queue{
		Songs: make([]*song.Song, 0),
	}
}

func (q *Queue) Len() int {
	return len(q.Songs)
}

func (q *Queue) Less(i, j int) bool {
	if len(q.Songs[i].GetVotes()) != len(q.Songs[j].GetVotes()) {
		return len(q.Songs[i].GetVotes()) > len(q.Songs[j].GetVotes())
	}
	return q.Songs[i].Date.Before(q.Songs[j].Date)
}

func (q *Queue) Swap(i, j int) {
	q.Songs[i], q.Songs[j] = q.Songs[j], q.Songs[i]
}

func (q *Queue) Tracks() []types.Song {
	songs := make([]types.Song, len(q.Songs))
	for i, s := range q.Songs {
		songs[i] = types.Song{
			Song:  s.Track,
			Votes: len(s.GetVotes()),
			Voted: false,
		}
	}
	return songs
}

func (q *Queue) For(voteID string) []types.Song {
	songs := make([]types.Song, 0)
	for _, s := range q.Songs {
		songs = append(songs, types.Song{
			Song:  s.Track,
			Votes: len(s.GetVotes()),
			Voted: s.HasVote(voteID),
		})
	}
	return songs
}

func (q *Queue) GetNext() (*song.Song, error) {
	if len(q.Songs) == 0 {
		return nil, ErrQueueEmpty
	}
	var s *song.Song
	s = q.Songs[0]
	return s, nil
}

func (q *Queue) GetHistory(voteID string) []types.Song {
	songs := make([]types.Song, 0)
	for _, s := range q.History {
		songs = append(songs, types.Song{
			Song:  s.Track,
			Votes: len(s.GetVotes()),
			Voted: s.HasVote(voteID),
		})
	}
	return songs
}

func (q *Queue) Vote(songID string, voteID string, s interface{}) error {
	if q.containsSong(songID) {
		so := q.Songs[q.indexOf(songID)]
		so.Vote(voteID)
	} else {
		so, err := q.add(s)
		if err != nil {
			return err
		}
		if voteID == HostVoteIdentifier {
			so.Date = so.Date.Add(time.Hour * 24 * 365)
		}
		so.Vote(voteID)
	}

	q.removeEmptySongs()
	sort.Sort(q)
	return nil
}

func (q *Queue) Advance() error {
	if len(q.Songs) == 0 {
		return ErrQueueEmpty
	}
	q.History = append(q.History, q.Songs[0])
	q.Songs = append(q.Songs[:0], q.Songs[1:]...)
	return nil
}

func (q *Queue) Delete(songID string) error {
	if !q.containsSong(songID) {
		return ErrSongNotFound
	}
	index := q.indexOf(songID)
	q.Songs = append(q.Songs[:index], q.Songs[index+1:]...)
	return nil
}

func (q *Queue) add(s interface{}) (*song.Song, error) {
	so := song.New(s.(*spotify.FullTrack))
	q.Songs = append(q.Songs, so)
	return so, nil
}

func (q *Queue) containsSong(songID string) bool {
	for _, s := range q.Songs {
		if s.ID == songID {
			return true
		}
	}
	return false
}

func (q *Queue) indexOf(songID string) int {
	for i, s := range q.Songs {
		if s.ID == songID {
			return i
		}
	}
	return -1
}

func (q *Queue) removeEmptySongs() {
	for _, s := range q.Songs {
		if len(s.GetVotes()) == 0 {
			_ = q.Delete(s.ID)
		}
	}
}
