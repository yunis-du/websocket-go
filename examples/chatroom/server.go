package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/yunis-du/websocket-go/http"
	"github.com/yunis-du/websocket-go/websocket"
)

// ChatRoom manages all connected clients
type ChatRoom struct {
	clients map[websocket.Speaker]string
	mu      sync.RWMutex
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{
		clients: make(map[websocket.Speaker]string),
	}
}

func (cr *ChatRoom) AddClient(c websocket.Speaker, username string) {
	cr.mu.Lock()
	cr.clients[c] = username
	cr.mu.Unlock()
}

func (cr *ChatRoom) RemoveClient(c websocket.Speaker) {
	cr.mu.Lock()
	delete(cr.clients, c)
	cr.mu.Unlock()
}

func (cr *ChatRoom) Broadcast(msg string, sender websocket.Speaker) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	for client := range cr.clients {
		if client != sender {
			_ = client.SendRaw([]byte(msg))
		}
	}
}

func (cr *ChatRoom) GetUsername(c websocket.Speaker) string {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.clients[c]
}

func (cr *ChatRoom) GetOnlineCount() int {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return len(cr.clients)
}

// ChatMessage represents a chat message structure
type ChatMessage struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	Type     string `json:"type"` // join, leave, message
}

type chatEventHandler struct {
	room *ChatRoom
}

func (h *chatEventHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.JSONProtocol)
	username := h.room.GetUsername(c)

	msg := fmt.Sprintf("[%s]: %s", username, string(bytes))
	log.Printf("Received message: %s", msg)

	// Broadcast to other clients
	h.room.Broadcast(msg, c)
}

func (h *chatEventHandler) OnConnected(c websocket.Speaker) {
	// Get username from Headers
	username := c.GetHeaders().Get("X-Username")
	if username == "" {
		username = fmt.Sprintf("User-%s", c.GetHost())
	}

	h.room.AddClient(c, username)
	log.Printf("User [%s] joined chatroom, online: %d", username, h.room.GetOnlineCount())

	// Notify everyone
	joinMsg := fmt.Sprintf("System: [%s] joined the chatroom", username)
	h.room.Broadcast(joinMsg, nil)
}

func (h *chatEventHandler) OnDisconnected(c websocket.Speaker) {
	username := h.room.GetUsername(c)
	h.room.RemoveClient(c)

	log.Printf("User [%s] left chatroom, online: %d", username, h.room.GetOnlineCount())

	// Notify everyone
	leaveMsg := fmt.Sprintf("System: [%s] left the chatroom", username)
	h.room.Broadcast(leaveMsg, nil)
}

func main() {
	hostname, _ := os.Hostname()
	httpServer := http.NewServer(hostname, "0.0.0.0", 8080)
	httpServer.ListenAndServe()

	wsServer := websocket.NewServer(httpServer, "/ws/chat")

	// Create chatroom
	room := NewChatRoom()
	wsServer.AddEventHandler(&chatEventHandler{room: room})

	wsServer.Start()

	log.Println("Chatroom server started at ws://0.0.0.0:8080/ws/chat")
	log.Println("Waiting for client connections...")

	select {}
}
