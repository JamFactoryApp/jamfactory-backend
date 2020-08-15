package types

import (
	"github.com/zmb3/spotify"
	"jamfactory-backend/models"
)

// ---------------------------------------------------------------------------------------------------------------------
// general

type JamResponse struct {
	Label    string     `json:"label"`
	Name     string     `json:"name"`
	Active   bool       `json:"active"`
	IpVoting bool       `json:"ip_voting"`
}

type PlaybackBody struct {
	Playback spotify.PlayerState `json:"playback"`
	DeviceID spotify.ID `json:"device_id"`
}

type LabelResponse struct {
	Label string `json:"label"`
}

// ---------------------------------------------------------------------------------------------------------------------
// auth controller

type AuthCurrentResponse struct {
	User       string `json:"user"`
	Label      string `json:"label"`
	Authorized bool   `json:"authorized"`
}

type AuthLoginResponse struct {
	Url string `json:"url"`
}

type AuthLogoutResponse struct {
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
	Queue []models.SongWithoutId `json:"queue"`
}

type PutQueuePlaylistsResponse GetQueueResponse
type PutQueueVoteResponse GetQueueResponse

// ---------------------------------------------------------------------------------------------------------------------
// spotify controller

type GetSpotifyPlaylistsResponse struct {
	Playlists *spotify.SimplePlaylistPage `json:"playlists"`
}

type GetSpotifyDevicesResponse struct {
	Devices []spotify.PlayerDevice `json:"devices"`
}

type PutSpotifySearchResponse struct {
	SearchResult *spotify.SearchResult `json:"search_result"`
}
