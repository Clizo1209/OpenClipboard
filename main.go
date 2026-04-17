package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed public
var publicFS embed.FS

// Message represents a WebSocket message
type Message struct {
	Type    string `json:"type"`
	Room    string `json:"room,omitempty"`
	Content string `json:"content,omitempty"`
	Users   int    `json:"users,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	conn *websocket.Conn
	room string
	mu   sync.Mutex
}

// Room holds the state for a single room
type Room struct {
	clients map[*Client]bool
	content string
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	rooms   = make(map[string]*Room)
	roomsMu sync.RWMutex
)

func generateRoomID() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func getRoom(roomID string) *Room {
	roomsMu.RLock()
	defer roomsMu.RUnlock()
	return rooms[roomID]
}

func createRoom(roomID string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	r := &Room{
		clients: make(map[*Client]bool),
		content: "",
	}
	rooms[roomID] = r
	return r
}

func removeClientFromRoom(client *Client) {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	roomID := client.room
	r, ok := rooms[roomID]
	if !ok {
		return
	}
	delete(r.clients, client)

	if len(r.clients) == 0 {
		delete(rooms, roomID)
		log.Printf("Room %s dissolved (no users left)", roomID)
	} else {
		broadcastUsers(r)
	}
}

func broadcastUsers(r *Room) {
	msg := Message{Type: "users", Users: len(r.clients)}
	data, _ := json.Marshal(msg)
	for c := range r.clients {
		c.mu.Lock()
		c.conn.WriteMessage(websocket.TextMessage, data)
		c.mu.Unlock()
	}
}

func broadcastSync(r *Room) {
	msg := Message{Type: "sync", Content: r.content, Users: len(r.clients)}
	data, _ := json.Marshal(msg)
	for c := range r.clients {
		c.mu.Lock()
		c.conn.WriteMessage(websocket.TextMessage, data)
		c.mu.Unlock()
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := &Client{conn: conn}
	defer func() {
		if client.room != "" {
			removeClientFromRoom(client)
		}
		conn.Close()
	}()

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "join":
			roomID := strings.ToUpper(strings.TrimSpace(msg.Room))

			// Reject joining the same room
			if roomID != "" && roomID == client.room {
				errMsg := Message{Type: "error", Error: "你已在房间 " + roomID + " 中"}
				data, _ := json.Marshal(errMsg)
				client.mu.Lock()
				client.conn.WriteMessage(websocket.TextMessage, data)
				client.mu.Unlock()
				continue
			}

			// Leave old room if any
			if client.room != "" {
				removeClientFromRoom(client)
			}

			var room *Room
			if roomID == "" {
				// Create a new room
				roomID = generateRoomID()
				room = createRoom(roomID)
			} else {
				// Join existing room only
				room = getRoom(roomID)
				if room == nil {
					errMsg := Message{Type: "error", Error: "房间 " + roomID + " 不存在"}
					data, _ := json.Marshal(errMsg)
					client.mu.Lock()
					client.conn.WriteMessage(websocket.TextMessage, data)
					client.mu.Unlock()
					continue
				}
			}

			client.room = roomID

			roomsMu.Lock()
			room.clients[client] = true
			roomsMu.Unlock()

			// Send current state to the joining client
			syncMsg := Message{Type: "sync", Room: roomID, Content: room.content, Users: len(room.clients)}
			data, _ := json.Marshal(syncMsg)
			client.mu.Lock()
			client.conn.WriteMessage(websocket.TextMessage, data)
			client.mu.Unlock()

			// Notify others about user count change
			roomsMu.RLock()
			broadcastUsers(room)
			roomsMu.RUnlock()

			log.Printf("Client joined room %s (%d users)", roomID, len(room.clients))

		case "update":
			if client.room == "" {
				continue
			}
			content := msg.Content
			// Enforce max length
			runes := []rune(content)
			if len(runes) > 10000 {
				runes = runes[:10000]
				content = string(runes)
			}

			roomsMu.RLock()
			room, ok := rooms[client.room]
			roomsMu.RUnlock()
			if !ok {
				continue
			}

			roomsMu.Lock()
			room.content = content
			roomsMu.Unlock()

			roomsMu.RLock()
			broadcastSync(room)
			roomsMu.RUnlock()
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	sub, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/", http.FileServer(http.FS(sub)))

	port := flag.Int("port", 8080, "listening port")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("OpenClipboard server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
