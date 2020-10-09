package notifications

import (
	"bytes"
	"encoding/gob"
)

type Message struct {
	Event   WebsocketEvent
	Message interface{}
}

func (m *Message) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(m)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (m *Message) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(&m)
}

type WebsocketEvent string

const (
	Playback WebsocketEvent = "playback"
	Queue                   = "queue"
	Close                   = "close"
)

type WebsocketCloseType string

const (
	HostLeft           WebsocketCloseType = "host"
	JamSessionInactive                    = "inactive"
)
