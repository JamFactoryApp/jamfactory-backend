package types

import (
	"encoding/gob"
	"golang.org/x/oauth2"
)

func RegisterGobTypes() {
	gob.Register(&oauth2.Token{})
}
