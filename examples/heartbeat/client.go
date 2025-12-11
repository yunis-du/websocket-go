package main

import (
	"log"
	"net/url"
	"os"
	"time"

	"github.com/yunis-du/websocket-go/websocket"
)

type heartbeatClientHandler struct {
	lastAck time.Time
}

func (h *heartbeatClientHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.RawProtocol)
	msg := string(bytes)

	if msg == "heartbeat_ack" {
		h.lastAck = time.Now()
		log.Println("💓 Heartbeat ack received")
	} else {
		log.Printf("Received message: %s", msg)
	}
}

func (h *heartbeatClientHandler) OnConnected(c websocket.Speaker) {
	log.Println("✓ Connected to server")
	h.lastAck = time.Now()
}

func (h *heartbeatClientHandler) OnDisconnected(c websocket.Speaker) {
	log.Println("✗ Disconnected")
}

func main() {
	hostname, _ := os.Hostname()

	u := &url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/heartbeat"}
	opt := websocket.ClientOpts{
		Protocol:  websocket.RawProtocol,
		QueueSize: 1000,
	}

	wsClient := websocket.NewClient(hostname, u, opt)
	handler := &heartbeatClientHandler{}
	wsClient.AddEventHandler(handler)

	wsClient.Start()

	// Wait for connection
	time.Sleep(500 * time.Millisecond)

	// 启动心跳send器
	heartbeatInterval := 5 * time.Second
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	log.Printf("Start sending heartbeat, interval: %v", heartbeatInterval)

	// Statistics
	sentCount := 0
	startTime := time.Now()

	for range ticker.C {
		if !wsClient.IsConnected() {
			log.Println("Connection closed, stop sending heartbeat")
			return
		}

		err := wsClient.SendRaw([]byte("heartbeat"))
		if err != nil {
			log.Printf("❌ Send heartbeat failed: %v", err)
		} else {
			sentCount++
			elapsed := time.Since(startTime)
			log.Printf("💓 Send heartbeat #%d (uptime: %v)", sentCount, elapsed.Round(time.Second))
		}

		// Check heartbeat response timeout
		if time.Since(handler.lastAck) > 20*time.Second {
			log.Println("⚠️  Heartbeat response timeout, connection may be abnormal")
		}
	}
}
