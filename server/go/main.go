package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	room string
}

var (
	upgrader   = websocket.Upgrader{}
	rooms      = make(map[string][]*Client)
	roomsMutex = sync.Mutex{}
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("../../client")))
	http.HandleFunc("/signal", handleSignal)

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleSignal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	var client *Client

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		roomID := msg["room"].(string)

		switch msg["type"] {
		case "join":
			client = &Client{conn: conn, room: roomID}
			roomsMutex.Lock()
			rooms[roomID] = append(rooms[roomID], client)
			roomsMutex.Unlock()
			log.Printf("Client joined room %s\n", roomID)

			if len(rooms[roomID]) == 2 {
				log.Println("Room is full, notifying peers to connect")
				msg := map[string]interface{}{
					"type": "ready",
					"room": roomID,
				}
				// notify both clients
				for _, client := range rooms[roomID] {
					client.conn.WriteJSON(msg)
				}
			}
		case "offer", "answer", "candidate":
			broadcastToRoom(roomID, msg, conn)
		}
	}

	// Remove client on disconnect
	if client != nil {
		roomsMutex.Lock()
		conns := rooms[client.room]
		for i, c := range conns {
			if c.conn == conn {
				rooms[client.room] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		roomsMutex.Unlock()
	}
}

func broadcastToRoom(room string, msg map[string]interface{}, sender *websocket.Conn) {
	roomsMutex.Lock()
	defer roomsMutex.Unlock()

	for _, client := range rooms[room] {
		if client.conn != sender {
			msg["room"] = room // make sure it's set
			if err := client.conn.WriteJSON(msg); err != nil {
				log.Println("Write error:", err)
			}
		}
	}
}
