package models

import (
	"github.com/zmb3/spotify"
	"sort"
	"time"
)

type songs []Song

func (song songs) Len() int {
	return len(song)
}

func (song songs) Swap(i, j int) {
	song[i], song[j] = song[j], song[i]
}

func (song songs) Less(i, j int) bool {
	if song[i].GetVoteCount() > song[j].GetVoteCount() {
		return true
	}
	if song[i].GetVoteCount() < song[j].GetVoteCount() {
		return false
	}
	if song[j].Date.After(song[i].Date) {
		return true
	}
	if song[i].Date.After(song[j].Date) {
		return false
	}
	return true
}

type Queue struct {
	Songs songs
}

func (queue *Queue) Vote(id string, song spotify.FullTrack) {
	notInQueueFlag := true

	for _, a := range queue.Songs {
		if a.Song.URI == song.URI {
			a.Vote(id)
			notInQueueFlag = false
		}
	}

	if notInQueueFlag {
		song := Song{Song:  song}
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

func (queue *Queue) CheckForEmptySongs() {
	for i, a := range queue.Songs {
		if a.GetVoteCount() <= 0 {
			queue.Songs = append(queue.Songs[:i], queue.Songs[i+1:]...)
		}
	}
}

func (queue *Queue) SortQueue() {
	sort.Sort(songs(queue.Songs))
}

func (queue *Queue) GetNextSong(removeSong bool) Song {
	song := queue.Songs[0]
	if removeSong {
		queue.Songs = queue.Songs[1:]
	}
	return song
}

func (queue *Queue) GetObjectWithoutId(id string) []SongWithoutId {
	res := make([]SongWithoutId, len(queue.Songs))

	for i, song := range queue.Songs {
		res[i] = song.GetObjectWithoutId(id)
	}

	return res
}