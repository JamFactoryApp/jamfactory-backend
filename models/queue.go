package models

import (
	"errors"
	"github.com/zmb3/spotify"
	"sort"
	"time"
)

type PartyQueue struct {
	Songs  []Song
}

func (queue *PartyQueue) Len() int {
	return len(queue.Songs)
}

func (queue *PartyQueue) Swap(i, j int) {
	queue.Songs[i], queue.Songs[j] = queue.Songs[j], queue.Songs[i]
}

func (queue *PartyQueue) Less(i, j int) bool {
	if queue.Songs[i].VoteCount() != queue.Songs[j].VoteCount() {
		return queue.Songs[i].VoteCount() > queue.Songs[j].VoteCount()
	} else {
		return queue.Songs[j].Date.Before(queue.Songs[i].Date)
	}
}

func (queue *PartyQueue) Vote(id string, song spotify.FullTrack) {
	notInQueueFlag := true

	for _, a := range queue.Songs {
		if a.Song.URI == song.URI {
			a.Vote(id)
			notInQueueFlag = false
		}
	}

	if notInQueueFlag {
		song := Song{Song: song}
		song.Vote(id)
		song.Date = time.Now()
		if id == "Host" {
			song.Date.Add(time.Hour * 24 * 365)
		}
		queue.Songs = append(queue.Songs, song)
	}

	queue.CheckForEmptySongs()
	queue.SortQueue()
}

func (queue *PartyQueue) CheckForEmptySongs() {
	for i, a := range queue.Songs {
		if a.VoteCount() <= 0 {
			queue.Songs = append(queue.Songs[:i], queue.Songs[i+1:]...)
		}
	}
}

func (queue *PartyQueue) SortQueue() {
	sort.Sort(queue)
}

func (queue *PartyQueue) GetNextSong(removeSong bool) (*Song, error) {
	if len(queue.Songs) == 0 {
		return nil, errors.New("no song")
	}
	song := queue.Songs[0]
	if removeSong {
		queue.Songs = queue.Songs[1:]
	}
	return &song, nil
}

func (queue *PartyQueue) GetObjectWithoutId(id string) []SongWithoutId {
	res := make([]SongWithoutId, len(queue.Songs))
	for i, song := range queue.Songs {
		res[i] = song.WithoutId(id)
	}

	return res
}
