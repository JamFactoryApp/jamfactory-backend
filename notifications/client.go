package notifications

import (
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	Room *Room
	Conn *websocket.Conn
	Send chan *Message
}

func NewClient(room *Room, conn *websocket.Conn) *Client {
	return &Client{
		Room: room,
		Conn: conn,
		Send: make(chan *Message, 1),
	}
}

func (c *Client) Read() {
	defer func() {
		c.Room.Unregister <- c
		if err := c.Conn.Close(); err != nil {
			log.Trace("Error closing connection: ", err)
		}
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Trace("Error setting read deadline: ", err)
	}
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Trace("Error setting read deadline: ", err)
		}
		return nil
	})
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Trace("Unexpected connection close: ", err)
			}
			break
		}
		message := &Message{}
		if err := message.Deserialize(data); err != nil {
			log.Error("Failed to deserialize message: ", err)
			break
		}
		c.Room.Broadcast <- message
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.Conn.Close(); err != nil {
			log.Trace("Error closing connection: ", err)
		}
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Trace("Error setting write deadline: ", err)
			}
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Trace("Error writing message: ", err)
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			data, err := message.Serialize()
			if err != nil {
				log.Error("Failed to deserialize message: ", err)
				break
			}

			if _, err := w.Write(data); err != nil {
				log.Error("Error writing message: ", err)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Trace("Error setting write deadline: ", err)
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
