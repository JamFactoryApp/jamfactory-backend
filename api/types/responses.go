package types

import (
	"github.com/zmb3/spotify"
)

// ---------------------------------------------------------------------------------------------------------------------
// general

type JamResponse struct {
	Label  string `json:"label"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

type JamMember struct {
	DisplayName string       `json:"display_name"`
	Identifier  string       `json:"identifier"`
	Permission  []Permission `json:"permissions"`
}

type JamMemberResponse struct {
	Members []JamMember `json:"members"`
}

type PlaybackBody struct {
	Playback *spotify.PlayerState `json:"playback"`
	DeviceID spotify.ID           `json:"device_id"`
}

type LabelResponse struct {
	Label string `json:"label"`
}

type SuccessResponse struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------------------------------------------------
// auth controller

type GetAuthCurrentResponse struct {
	UserType   string `json:"user"`
	Label      string `json:"label"`
	Authorized bool   `json:"authorized"`
}

type GetAuthLoginResponse struct {
	URL string `json:"url"`
}

type GetAuthLogoutResponse SuccessResponse

// ---------------------------------------------------------------------------------------------------------------------
//
// controller

type GetJamResponse JamResponse
type PutJamResponse JamResponse

type GetJamPlaybackResponse PlaybackBody
type PutJamPlaybackResponse PlaybackBody

type GetJamMembersResponse JamMemberResponse
type PutJamMembersResponse JamMemberResponse

type GetJamCreateResponse LabelResponse
type PutJamJoinResponse LabelResponse

type GetJamLeaveResponse struct {
	Success bool `json:"success"`
}

// ---------------------------------------------------------------------------------------------------------------------
// queue controller

type GetQueueResponse struct {
	Tracks []Song `json:"tracks"`
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

// ---------------------------------------------------------------------------------------------------------------------
// user controller

type UserResponse struct {
	Identifier        string `json:"identifier"`
	DisplayName       string `json:"display_name"`
	UserType          string `json:"type"`
	JoinedLabel       string `json:"joined_label"`
	SpotifyAuthorized bool   `json:"spotify_authorized"`
}

type GetUserResponse UserResponse
type PutUserResponse UserResponse
type DeleteUserResponse SuccessResponse
