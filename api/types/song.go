package types

type Song struct {
	Song  interface{} `json:"spotifyTrackFull"`
	Votes int         `json:"votes"`
	Voted bool        `json:"voted"`
}
