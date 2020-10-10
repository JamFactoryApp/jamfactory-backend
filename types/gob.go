package types

import (
	"encoding/gob"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

func RegisterGobTypes() {
	gob.Register(&oauth2.Token{})
	gob.Register(&spotify.SearchResult{})
}
