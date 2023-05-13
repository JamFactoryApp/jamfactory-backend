package types

import "github.com/zmb3/spotify/v2"

type Song struct {
	Song  *spotify.FullTrack `json:"spotifyTrackFull"`
	Votes int                `json:"votes"`
	Voted bool               `json:"voted"`
}
