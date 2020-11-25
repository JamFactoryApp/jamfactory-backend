package types

import "github.com/zmb3/spotify"

type SocketJamState struct {
	CurrentSong *spotify.FullTrack   `json:"currentSong"`
	State       *spotify.PlayerState `json:"state"`
}

type SocketPlaybackState struct {
	Playback *spotify.PlayerState `json:"playback"`
}
