package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/yunis-du/websocket-go/websocket"
)

// ClientMonitor monitors client connections
type ClientMonitor struct {
	clients     map[websocket.Speaker]*ClientInfo
	mu          sync.RWMutex
	timeout     time.Duration
	checkTicker *time.Ticker
}

type ClientInfo struct {
	ConnectedAt    time.Time
	LastHeartbeat  time.Time
	HeartbeatCount int
	IsAlive        bool
}

func NewClientMonitor(timeout time.Duration) *ClientMonitor {
	return &ClientMonitor{
		clients: make(map[websocket.Speaker]*ClientInfo),
		timeout: timeout,
	}
}

func (cm *ClientMonitor) AddClient(c websocket.Speaker) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	cm.clients[c] = &ClientInfo{
		ConnectedAt:   now,
		LastHeartbeat: now,
		IsAlive:       true,
	}
}

func (cm *ClientMonitor) RemoveClient(c websocket.Speaker) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, c)
}

func (cm *ClientMonitor) UpdateHeartbeat(c websocket.Speaker) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if info, ok := cm.clients[c]; ok {
		info.LastHeartbeat = time.Now()
		info.HeartbeatCount++
		info.IsAlive = true
	}
}

func (cm *ClientMonitor) StartMonitoring() {
	cm.checkTicker = time.NewTicker(5 * time.Second)
	go func() {
		for range cm.checkTicker.C {
			cm.checkClients()
		}
	}()
}

func (cm *ClientMonitor) checkClients() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for client, info := range cm.clients {
		elapsed := now.Sub(info.LastHeartbeat)
		if elapsed > cm.timeout {
			if info.IsAlive {
				log.Printf("⚠️  Client %s heartbeat timeout (%.1fs), marked as offline", client.GetHost(), elapsed.Seconds())
				info.IsAlive = false
			}
		}
	}
}

func (cm *ClientMonitor) GetStats() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	total := len(cm.clients)
	alive := 0
	for _, info := range cm.clients {
		if info.IsAlive {
			alive++
		}
	}

	return fmt.Sprintf("Total connections: %d, Alive: %d, Offline: %d", total, alive, total-alive)
}

type heartbeatHandler struct {
	monitor *ClientMonitor
}

func (h *heartbeatHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.RawProtocol)
	msg := string(bytes)

	if msg == "heartbeat" {
		h.monitor.UpdateHeartbeat(c)
		// 回复心跳确认
		_ = c.SendRaw([]byte("heartbeat_ack"))
	} else {
		log.Printf("Received message: %s", msg)
	}
}

func (h *heartbeatHandler) OnConnected(c websocket.Speaker) {
	h.monitor.AddClient(c)
	log.Printf("✓ Client connected: %s", c.GetHost())
}

func (h *heartbeatHandler) OnDisconnected(c websocket.Speaker) {
	h.monitor.RemoveClient(c)
	log.Printf("✗ Client disconnected: %s", c.GetHost())
}

func main() {
	hostname, _ := os.Hostname()

	wsServer := websocket.NewServer(hostname, "0.0.0.0", 8080, "/ws/heartbeat")

	// Create monitor, timeout 15 seconds
	monitor := NewClientMonitor(15 * time.Second)
	monitor.StartMonitoring()

	wsServer.AddEventHandler(&heartbeatHandler{monitor: monitor})
	wsServer.Start()

	log.Println("Heartbeat server started at ws://0.0.0.0:8080/ws/heartbeat")
	log.Println("Heartbeat timeout: 15 seconds")

	// Periodically print statistics
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			log.Printf("📊 %s", monitor.GetStats())
		}
	}()

	select {}
}
