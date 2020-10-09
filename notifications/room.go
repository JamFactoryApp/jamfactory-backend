package notifications

import (
	log "github.com/sirupsen/logrus"
)

type Room struct {
	Clients    map[*Client]bool
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
	log        *log.Entry
}

func NewRoom() *Room {
	return &Room{
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (r *Room) OpenDoors() {
	for {
		select {
		case client := <-r.Register:
			log.Trace("Registered client: ", client)
			r.Clients[client] = true
		case client := <-r.Unregister:
			log.Trace("Unregistered client: ", client)
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
			}
		case message := <-r.Broadcast:
			log.Trace("Broadcasting message: ", message)
			for client := range r.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(r.Clients, client)
				}
			}
		}
	}
}
