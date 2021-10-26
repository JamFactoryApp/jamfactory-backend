package types

// ---------------------------------------------------------------------------------------------------------------------
// general

type JoinRequest struct {
	Label    string `json:"label"`
	Password string `json:"password"`
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

type PutQueueExportRequest struct {
	PlaylistName   string `json:"playlist_name"`
	IncludeHistory bool   `json:"include_history"`
	IncludeQueue   bool   `json:"include_queue"`
}

type DeleteQueueSongRequest struct {
	TrackID string `json:"track"`
}

// ---------------------------------------------------------------------------------------------------------------------
// jamsession controller

type PutJamRequest struct {
	Name     JSONString `json:"name,omitempty"`
	Active   JSONBool   `json:"active,omitempty"`
	Password JSONString `json:"password,omitempty"`
}

type PutJamPlaybackRequest struct {
	Playing  JSONBool   `json:"playing,omitempty"`
	DeviceID JSONString `json:"device_id,omitempty"`
}

type JamMemberRequest struct {
	Members []JamMember `json:"members"`
}

type PutJamMemberRequest JamMemberRequest

type PutJamJoinRequest JoinRequest

// ---------------------------------------------------------------------------------------------------------------------
// user controller

type UserRequest struct {
	DisplayName string `json:"display_name"`
}

type PutUserRequest UserRequest
