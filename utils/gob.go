package utils

import (
	"encoding/gob"
	"golang.org/x/oauth2"
)

func registerGobTypes() {
	gob.Register(&oauth2.Token{})
}
