package types

import "github.com/zmb3/spotify"

// ---------------------------------------------------------------------------------------------------------------------
// auth controller
type StatusResponseBody struct {
	User       string `json:"user"`
	Label      string `json:"label"`
	Authorized bool   `json:"authorized"`
}

type LoginResponseBody struct {
	Url string `json:"url"`
}
// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller

type JamSessionBody struct {
	Label    string     `json:"label"`
	Name     string     `json:"name"`
	Active   bool       `json:"active"`
	DeviceID spotify.ID 	`json:"device_id"`
	IpVoting bool       `json:"ip_voting"`
}
type GetJamSessionResponseBody JamSessionBody
type SetJamSessionResponseBody JamSessionBody

type PlaybackBody struct {
	Playback spotify.PlayerState `json:"playback"`
}
type GetPlaybackResponseBody PlaybackBody
type SetPlaybackResponseBody PlaybackBody

type labelBody struct {
	Label string `json:"label"`
}
type CreateJamSessionResponseBody labelBody
type JoinRequestBody labelBody
type JoinResponseBody labelBody

type LeaveJamSessionResponseBody struct {
	Success bool `json:"success"`
}

type JamSessionStateResponseBody struct {
	CurrentSong interface{} `json:"currentSong"`
	State       interface{} `json:"state"`
}