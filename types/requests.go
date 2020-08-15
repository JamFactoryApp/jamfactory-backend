package types

// ---------------------------------------------------------------------------------------------------------------------
// general

type LabelRequest struct {
	Label string `json:"label"`
}

// ---------------------------------------------------------------------------------------------------------------------
// spotify controller

type PutSpotifySearchRequest struct {
	SearchText JSONString `json:"text"`
	SearchType JSONString `json:"type"`
}

// ---------------------------------------------------------------------------------------------------------------------
// queue controller

type PutQueueVoteRequest struct {
	TrackID JSONString `json:"track"`
}

type PutQueuePlaylistRequest struct {
	PlaylistID JSONString `json:"playlist"`
}

// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller

type PutJamRequest struct {
	Name     JSONString `json:"name"`
	Active   JSONBool   `json:"active"`
	IpVoting JSONBool   `json:"ip_voting"`
}

type PutJamPlaybackRequest struct {
	Playing  JSONBool   `json:"playing,omitempty"`
	DeviceID JSONString `json:"device_id,omitempty"`
}

type PutJamJoinRequest LabelRequest
