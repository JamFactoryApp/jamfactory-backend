package queue

import (
	"github.com/jamfactoryapp/jamfactory-backend/api/types"
	"github.com/jamfactoryapp/jamfactory-backend/pkg/song"
	"github.com/zmb3/spotify"
	"sort"
	"time"
)

type SpotifyQueue struct {
	songs []song.Song
}

func NewSpotify() *SpotifyQueue {
	return &SpotifyQueue{
		songs: make([]song.Song, 0),
	}
}

func (q *SpotifyQueue) Len() int {
	return len(q.songs)
}

func (q *SpotifyQueue) Less(i, j int) bool {
	if len(q.songs[i].Votes()) != len(q.songs[j].Votes()) {
		return len(q.songs[i].Votes()) > len(q.songs[j].Votes())
	}
	return q.songs[i].Date().Before(q.songs[j].Date())
}

func (q *SpotifyQueue) Swap(i, j int) {
	q.songs[i], q.songs[j] = q.songs[j], q.songs[i]
}

func (q *SpotifyQueue) Tracks() []types.Song {
	songs := make([]types.Song, len(q.songs))
	for i, s := range q.songs {
		songs[i] = types.Song{
			Song:  s.Song(),
			Votes: len(s.Votes()),
			Voted: false,
		}
	}
	return songs
}

func (q *SpotifyQueue) For(voteID string) []types.Song {
	songs := make([]types.Song, 0)
	for _, s := range q.songs {
		songs = append(songs, types.Song{
			Song:  s.Song(),
			Votes: len(s.Votes()),
			Voted: s.HasVote(voteID),
		})
	}
	return songs
}

func (q *SpotifyQueue) GetNext() (song.Song, error) {
	if len(q.songs) == 0 {
		return nil, ErrQueueEmpty
	}
	var s song.Song
	s = q.songs[0]
	return s, nil
}

func (q *SpotifyQueue) Vote(songID string, voteID string, s interface{}) error {
	if q.containsSong(songID) {
		so := q.songs[q.indexOf(songID)]
		so.Vote(voteID)
	} else {
		so, err := q.add(s)
		if err != nil {
			return err
		}
		if voteID == HostVoteIdentifier {
			so.SetDate(so.Date().Add(time.Hour * 24 * 365))
		}
		so.Vote(voteID)
	}

	q.removeEmptySongs()
	sort.Sort(q)
	return nil
}

func (q *SpotifyQueue) Advance() error {
	if len(q.songs) == 0 {
		return ErrQueueEmpty
	}
	q.songs = append(q.songs[:0], q.songs[1:]...)
	return nil
}

func (q *SpotifyQueue) Delete(songID string) error {
	if !q.containsSong(songID) {
		return ErrSongNotFound
	}
	index := q.indexOf(songID)
	q.songs = append(q.songs[:index], q.songs[index+1:]...)
	return nil
}

func (q *SpotifyQueue) add(s interface{}) (song.Song, error) {
	so := song.NewSpotify(s.(*spotify.FullTrack))
	q.songs = append(q.songs, so)
	return so, nil
}

func (q *SpotifyQueue) containsSong(songID string) bool {
	for _, s := range q.songs {
		if s.ID() == songID {
			return true
		}
	}
	return false
}

func (q *SpotifyQueue) indexOf(songID string) int {
	for i, s := range q.songs {
		if s.ID() == songID {
			return i
		}
	}
	return -1
}

func (q *SpotifyQueue) removeEmptySongs() {
	for _, s := range q.songs {
		if len(s.Votes()) == 0 {
			_ = q.Delete(s.ID())
		}
	}
}
