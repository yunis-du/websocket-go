package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/yunis-du/websocket-go/websocket"
)

type poolClientHandler struct {
	clientID string
	mu       sync.Mutex
	received int
}

func (h *poolClientHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	h.mu.Lock()
	h.received++
	count := h.received
	h.mu.Unlock()

	bytes, _ := m.Bytes(websocket.RawProtocol)
	log.Printf("[Client %s] Received response #%d: %s", h.clientID, count, string(bytes))
}

func (h *poolClientHandler) OnConnected(c websocket.Speaker) {
	log.Printf("[Client %s] ✓ Connected", h.clientID)
}

func (h *poolClientHandler) OnDisconnected(c websocket.Speaker) {
	log.Printf("[Client %s] ✗ Disconnected", h.clientID)
}

func main() {
	hostname, _ := os.Hostname()

	// Create a client pool
	pool := websocket.NewClientPool("test-pool")

	// Add multiple clients to the pool
	clientCount := 3
	for i := 1; i <= clientCount; i++ {
		clientID := fmt.Sprintf("client-%d", i)

		u := &url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/pool"}
		opt := websocket.ClientOpts{
			Protocol:  websocket.RawProtocol,
			QueueSize: 1000,
		}

		client := websocket.NewClient(fmt.Sprintf("%s-%s", hostname, clientID), u, opt)
		handler := &poolClientHandler{clientID: clientID}
		client.AddEventHandler(handler)

		err := pool.AddClient(client)
		if err != nil {
			log.Fatalf("Add client failed: %v", err)
		}

		log.Printf("Add client: %s", clientID)
	}

	// Connect all clients
	log.Println("\nConnecting all clients...")
	pool.ConnectAll()

	// Wait for connections to be established
	time.Sleep(1 * time.Second)

	// Test: Send messages using the client pool
	log.Println("\n=== Test 1: Polling send messages ===")
	for i := 1; i <= 9; i++ {
		msg := fmt.Sprintf("Message #%d", i)

		// Get a client from the pool (round-robin)
		speakers := pool.GetSpeakers()
		if len(speakers) > 0 {
			speaker := speakers[i%len(speakers)]
			err := speaker.SendRaw([]byte(msg))
			if err != nil {
				log.Printf("send failed: %v", err)
			} else {
				log.Printf("via %s send: %s", speaker.GetHost(), msg)
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(2 * time.Second)

	// Test: Broadcast message
	log.Println("\n=== Test 2: Broadcast message ===")
	broadcastMsg := "This is a broadcast message"
	for _, speaker := range pool.GetSpeakers() {
		err := speaker.SendRaw([]byte(broadcastMsg))
		if err != nil {
			log.Printf("broadcast failed: %v", err)
		}
	}

	time.Sleep(2 * time.Second)

	// Test: Disconnect one client
	log.Println("\n=== Test 3: Disconnect one client ===")
	speakers := pool.GetSpeakers()
	if len(speakers) > 0 {
		firstClient := speakers[0]
		log.Printf("Disconnect client: %s", firstClient.GetHost())
		firstClient.Stop()
	}

	time.Sleep(1 * time.Second)

	// Continue to send messages (using remaining clients)
	log.Println("\n=== Test 4: Send messages using remaining clients ===")
	for i := 1; i <= 5; i++ {
		msg := fmt.Sprintf("Message after disconnect #%d", i)
		speakers := pool.GetSpeakers()

		// Only use connected clients
		for _, speaker := range speakers {
			if speaker.IsConnected() {
				err := speaker.SendRaw([]byte(msg))
				if err != nil {
					log.Printf("send failed: %v", err)
				} else {
					log.Printf("via %s send: %s", speaker.GetHost(), msg)
					break
				}
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(2 * time.Second)

	// Close all connections
	log.Println("\nClosing all connections...")
	for _, speaker := range pool.GetSpeakers() {
		speaker.Stop()
	}

	time.Sleep(1 * time.Second)
	log.Println("Test completed")
}
