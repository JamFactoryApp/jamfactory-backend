package types

import (
	"github.com/zmb3/spotify"
)

// ---------------------------------------------------------------------------------------------------------------------
// general

type JamResponse struct {
	Label      string     `json:"label"`
	Name       string     `json:"name"`
	Active     bool       `json:"active"`
	VotingType VotingType `json:"voting_type"`
}

type PlaybackBody struct {
	Playback spotify.PlayerState `json:"playback"`
	DeviceID spotify.ID          `json:"device_id"`
}

type LabelResponse struct {
	Label string `json:"label"`
}

// ---------------------------------------------------------------------------------------------------------------------
// auth controller

type GetAuthCurrentResponse struct {
	User       string `json:"user"`
	Label      string `json:"label"`
	Authorized bool   `json:"authorized"`
}

type GetAuthLoginResponse struct {
	Url string `json:"url"`
}

type GetAuthLogoutResponse struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller

type GetJamResponse JamResponse
type PutJamResponse JamResponse

type GetJamPlaybackResponse PlaybackBody
type PutJamPlaybackResponse PlaybackBody

type GetJamCreateResponse LabelResponse
type PutJamJoinResponse LabelResponse

type GetJamLeaveResponse struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------------------------------------------------
// queue controller

type GetQueueResponse struct {
	Queue []SongWithoutId `json:"queue"`
}

type PutQueuePlaylistsResponse GetQueueResponse
type PutQueueVoteResponse GetQueueResponse
type DeleteQueueSongResponse GetQueueResponse

// ---------------------------------------------------------------------------------------------------------------------
// spotify controller

type GetSpotifyPlaylistsResponse struct {
	Playlists *spotify.SimplePlaylistPage `json:"playlists"`
}

type GetSpotifyDevicesResponse struct {
	Devices []spotify.PlayerDevice `json:"devices"`
}

type PutSpotifySearchResponse struct {
	Artists   *spotify.FullArtistPage     `json:"artists"`
	Albums    *spotify.SimpleAlbumPage    `json:"albums"`
	Playlists *spotify.SimplePlaylistPage `json:"playlists"`
	Tracks    *spotify.FullTrackPage      `json:"tracks"`
}
