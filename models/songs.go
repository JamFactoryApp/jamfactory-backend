package models

type Songs []Song

func (songs Songs) Len() int {
	return len(songs)
}

func (songs Songs) Swap(i, j int) {
	songs[i], songs[j] = songs[j], songs[i]
}

func (songs Songs) Less(i, j int) bool {
	if songs[i].VoteCount() != songs[j].VoteCount() {
		return songs[i].VoteCount() < songs[j].VoteCount()
	} else {
		return songs[i].Date.Before(songs[j].Date)
	}
}
