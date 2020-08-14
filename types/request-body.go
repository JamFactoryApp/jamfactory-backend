package types

// ---------------------------------------------------------------------------------------------------------------------
// spotify controller
type SearchRequestBody struct {
	SearchText JSONString `json:"text"`
	SearchType JSONString `json:"type"`
}

// ---------------------------------------------------------------------------------------------------------------------
// queue controller
type VoteRequestBody struct {
	TrackID JSONString `json:"track"`
}

type AddPlaylistRequestBody struct {
	PlaylistID JSONString `json:"playlist"`
}

// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller
type SetJamSessionRequestBody struct {
	Name     JSONString `json:"name"`
	Active   JSONBool   `json:"active"`
	IpVoting JSONBool   `json:"ip_voting"`
}

type SetPlayBackRequestBody struct {
	Playing  JSONBool   `json:"playing,omitempty"`
	DeviceID JSONString `json:"device_id,omitempty"`
}
