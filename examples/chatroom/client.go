package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/yunis-du/websocket-go/websocket"
)

type chatClientHandler struct {
	username string
}

func (h *chatClientHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.RawProtocol)
	fmt.Printf("\r%s\n> ", string(bytes))
}

func (h *chatClientHandler) OnConnected(c websocket.Speaker) {
	fmt.Printf("✓ Connected to chatroom, your username: %s\n", h.username)
	fmt.Println("Type message and press Enter to send, type /quit to exit")
	fmt.Print("> ")
}

func (h *chatClientHandler) OnDisconnected(c websocket.Speaker) {
	fmt.Println("\n✗ Disconnected")
	os.Exit(0)
}

func main() {
	// Get username
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "Anonymous"
	}

	// Connect to server
	// Set username in Header
	headers := http.Header{}
	headers.Set("X-Username", username)

	u := &url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/chat"}
	opt := websocket.ClientOpts{
		Protocol:  websocket.RawProtocol,
		QueueSize: 1000,
		Headers:   headers,
	}

	hostname, _ := os.Hostname()
	wsClient := websocket.NewClient(hostname, u, opt)

	handler := &chatClientHandler{username: username}
	wsClient.AddEventHandler(handler)

	wsClient.Start()

	// Read user input and send
	for {
		fmt.Print("> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "" {
			continue
		}

		if text == "/quit" {
			fmt.Println("Goodbye!")
			wsClient.Stop()
			break
		}

		err := wsClient.SendRaw([]byte(text))
		if err != nil {
			log.Printf("Send failed: %v", err)
		}
	}
}
