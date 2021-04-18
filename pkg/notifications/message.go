package notifications

import (
	"bytes"
	"encoding/json"
)

type Message struct {
	Event   WebsocketEvent `json:"event"`
	Message interface{}    `json:"message"`
}

func (m *Message) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	err := encoder.Encode(m)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (m *Message) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	decoder := json.NewDecoder(buffer)
	return decoder.Decode(&m)
}

type WebsocketEvent string

const (
	Playback WebsocketEvent = "playback"
	Queue                   = "queue"
	Close                   = "close"
	Jam                     = "jam"
)

type WebsocketCloseType string

const (
	HostLeft WebsocketCloseType = "host"
	Inactive                    = "inactive"
	Warning                     = "warning"
)
