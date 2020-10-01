package types

// ---------------------------------------------------------------------------------------------------------------------
// general

type LabelRequest struct {
	Label string `json:"label"`
}

// ---------------------------------------------------------------------------------------------------------------------
// spotify controller

type PutSpotifySearchRequest struct {
	SearchText string `json:"text"`
	SearchType string `json:"type"`
}

// ---------------------------------------------------------------------------------------------------------------------
// queue controller

type PutQueueVoteRequest struct {
	TrackID string `json:"track"`
}

type PutQueuePlaylistRequest struct {
	PlaylistID string `json:"playlist"`
}

type DeleteQueueSongRequest struct {
	TrackID string `json:"track"`
}

// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller

type PutJamRequest struct {
	Name     JSONString `json:"name,omitempty"`
	Active   JSONBool   `json:"active,omitempty"`
	VotingType JSONString   `json:"voting_type,omitempty"`
}

type PutJamPlaybackRequest struct {
	Playing  JSONBool   `json:"playing,omitempty"`
	DeviceID JSONString `json:"device_id,omitempty"`
}

type PutJamJoinRequest LabelRequest
