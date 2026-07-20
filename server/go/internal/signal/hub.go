package signal

import (
	"encoding/json"
	"log/slog"
	"sync"
)

type Hub struct {
	rooms      map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Route      chan *Message
	logger     *slog.Logger
	mu         sync.RWMutex
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Route:      make(chan *Message),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)
		case client := <-h.Unregister:
			h.handleUnregister(client)
		case msg := <-h.Route:
			h.handleRoute(msg)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[client.room] == nil {
		h.rooms[client.room] = make(map[*Client]bool)
	}

	h.rooms[client.room][client] = true

	peerCount := len(h.rooms[client.room])
	h.logger.Info("client joined room",
		"room", client.room,
		"client", client.id,
		"peers", peerCount,
	)

	if peerCount == 2 {
		for c := range h.rooms[client.room] {
			if c != client {
				c.send <- mustMarshal(&Message{
					Type: MsgReady,
					Room: client.room,
				})
			}
		}
	}

	h.broadcastPeerCount(client.room)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, ok := h.rooms[client.room]
	if !ok {
		return
	}

	if _, ok := clients[client]; !ok {
		return
	}

	delete(clients, client)
	close(client.send)

	h.logger.Info("client left room",
		"room", client.room,
		"client", client.id,
		"remaining", len(clients),
	)

	if len(clients) == 0 {
		delete(h.rooms, client.room)
		h.logger.Info("room deleted (empty)", "room", client.room)
	} else {
		for c := range clients {
			c.send <- mustMarshal(&Message{
				Type: MsgLeave,
				Room: client.room,
				Payload: mustMarshal(JoinPayload{
					Name: client.id,
				}),
			})
		}
		h.broadcastPeerCount(client.room)
	}
}

func (h *Hub) handleRoute(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[msg.Room]
	if !ok {
		return
	}

	data := mustMarshal(msg)

	for c := range clients {
		if c.id != msg.SenderID {
			select {
			case c.send <- data:
			default:
				h.logger.Warn("client send buffer full, dropping message",
					"client", c.id,
					"type", msg.Type,
				)
			}
		}
	}
}

func (h *Hub) broadcastPeerCount(room string) {
	clients, ok := h.rooms[room]
	if !ok {
		return
	}

	count := len(clients)
	data := mustMarshal(&PeerCountPayload{Count: count})
	msg := mustMarshal(&Message{
		Type:    MsgPeerCount,
		Room:    room,
		Payload: data,
	})

	for c := range clients {
		select {
		case c.send <- msg:
		default:
		}
	}
}

func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic("signal: json marshal failed: " + err.Error())
	}
	return data
}
