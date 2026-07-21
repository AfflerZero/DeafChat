package signal

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	rooms      map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Route      chan *Message
	logger     *slog.Logger
	maxRooms   int
	maxPerRoom int
	maxClients int
	mu         sync.RWMutex
}

func NewHub(logger *slog.Logger, maxRooms, maxPerRoom, maxClients int) *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Route:      make(chan *Message),
		logger:     logger,
		maxRooms:   maxRooms,
		maxPerRoom: maxPerRoom,
		maxClients: maxClients,
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

	if !h.canAcceptLocked(client.room) {
		h.logger.Warn("rejecting websocket client due to hub limits", "room", client.room, "client", client.id)
		close(client.send)
		_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "room capacity reached"))
		client.conn.Close()
		return
	}

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

func (h *Hub) CanAccept(room string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.canAcceptLocked(room)
}

func (h *Hub) canAcceptLocked(room string) bool {
	if h.maxRooms > 0 && h.rooms[room] == nil && len(h.rooms) >= h.maxRooms {
		return false
	}

	if h.maxPerRoom > 0 && len(h.rooms[room]) >= h.maxPerRoom {
		return false
	}

	if h.maxClients > 0 && h.totalClientsLocked() >= h.maxClients {
		return false
	}

	return true
}

func (h *Hub) totalClientsLocked() int {
	total := 0
	for _, clients := range h.rooms {
		total += len(clients)
	}
	return total
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
