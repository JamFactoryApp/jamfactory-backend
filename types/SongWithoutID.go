package types

import "github.com/zmb3/spotify"

type SongWithoutId struct {
	Song  *spotify.FullTrack `json:"spotifyTrackFull"`
	Votes int                `json:"votes"`
	Voted bool               `json:"voted"`
}
