package types

import "github.com/zmb3/spotify"

type SocketJamState struct {
	CurrentSong interface{} `json:"currentSong"`
	State       interface{} `json:"state"`
}

type SocketPlaybackState struct {
	Playback spotify.PlayerState `json:"playback"`
}
