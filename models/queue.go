package models

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/zmb3/spotify"
	"jamfactory-backend/types"
	"sort"
	"time"
)

type Queue struct {
	Songs []Song
}

func (queue *Queue) Len() int {
	return len(queue.Songs)
}

func (queue *Queue) Swap(i, j int) {
	queue.Songs[i], queue.Songs[j] = queue.Songs[j], queue.Songs[i]
}

func (queue *Queue) Less(i, j int) bool {
	if queue.Songs[i].VoteCount() != queue.Songs[j].VoteCount() {
		return queue.Songs[i].VoteCount() > queue.Songs[j].VoteCount()
	} else {
		return queue.Songs[i].Date.After(queue.Songs[j].Date)
	}
}

func (queue *Queue) Vote(id string, song *spotify.FullTrack) {
	notInQueueFlag := true

	for i := range queue.Songs {
		if queue.Songs[i].Song.ID == song.ID {
			var added = queue.Songs[i].Vote(id)
			notInQueueFlag = false
			log.WithFields(log.Fields{
				"Song":  queue.Songs[i].Song.Name,
				"Added": added}).Trace("Added Vote ", id)
		}
	}

	if notInQueueFlag {
		song := Song{Song: song}
		song.Vote(id)
		song.Date = time.Now()
		if id == UserTypeHost {
			song.Date.Add(time.Hour * 24 * 365)
		}
		queue.Songs = append(queue.Songs, song)
		log.WithField("Song", song.Song.Name).Trace("Added Song ", id)
	}

	queue.CheckForEmptySongs()
	queue.SortQueue()
}

func (queue *Queue) DeleteSong(id spotify.ID) bool {
	for i := range queue.Songs {
		if queue.Songs[i].Song.ID == id {
			queue.Songs = append(queue.Songs[:i], queue.Songs[i+1:]...)
			return true
		}
	}
	return false
}

func (queue *Queue) CheckForEmptySongs() {
	for i, a := range queue.Songs {
		if a.VoteCount() <= 0 {
			queue.Songs = append(queue.Songs[:i], queue.Songs[i+1:]...)
		}
	}
}

func (queue *Queue) SortQueue() {
	sort.Sort(queue)
}

func (queue *Queue) GetNextSong() (*Song, error) {
	if len(queue.Songs) == 0 {
		return nil, errors.New("no song")
	}
	song := queue.Songs[0]
	return &song, nil
}

func (queue *Queue) AdvanceQueue() error {
	if len(queue.Songs) == 0 {
		return errors.New("no song")
	}
	queue.Songs = queue.Songs[1:]
	return nil
}

func (queue *Queue) GetSong(id string) (*Song, error) {
	for _, song := range queue.Songs {
		if song.Song.ID.String() == id {
			return &song, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("song %s does not exist in queue", id))
}

func (queue *Queue) GetObjectWithoutId(id string) []types.SongWithoutId {
	res := make([]types.SongWithoutId, len(queue.Songs))
	for i, song := range queue.Songs {
		res[i] = song.WithoutId(id)
	}

	return res
}
