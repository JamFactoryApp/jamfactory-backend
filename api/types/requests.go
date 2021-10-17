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

type PutQueueCollectionRequest struct {
	CollectionID   string `json:"collection"`
	CollectionType string `json:"type"`
}

type DeleteQueueSongRequest struct {
	TrackID string `json:"track"`
}

// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller

type PutJamRequest struct {
	Name   JSONString `json:"name,omitempty"`
	Active JSONBool   `json:"active,omitempty"`
}

type PutJamPlaybackRequest struct {
	Playing  JSONBool   `json:"playing,omitempty"`
	DeviceID JSONString `json:"device_id,omitempty"`
}

type JamMemberRequest struct {
	Members []JamMember `json:"members"`
}

type PutJamMemberRequest JamMemberRequest

type PutJamJoinRequest LabelRequest

// ---------------------------------------------------------------------------------------------------------------------
// user controller

type UserRequest struct {
	DisplayName string `json:"display_name"`
}

type PutUserRequest UserRequest
