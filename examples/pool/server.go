package main

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	httpServer "github.com/yunis-du/websocket-go/http"
	"github.com/yunis-du/websocket-go/websocket"
)

type loadBalancerHandler struct {
	requestCount int64
}

func (h *loadBalancerHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.RawProtocol)

	// add request count
	count := atomic.AddInt64(&h.requestCount, 1)

	response := fmt.Sprintf("Server response #%d: Received message '%s'", count, string(bytes))
	log.Printf("Processing request #%d from %s", count, c.GetHost())

	_ = c.SendRaw([]byte(response))
}

func (h *loadBalancerHandler) OnConnected(c websocket.Speaker) {
	log.Printf("✓ Client connected: %s", c.GetHost())
}

func (h *loadBalancerHandler) OnDisconnected(c websocket.Speaker) {
	log.Printf("✗ Client disconnected: %s", c.GetHost())
}

func main() {
	hostname, _ := os.Hostname()
	httpSrv := httpServer.NewServer(hostname, "0.0.0.0", 8080)
	httpSrv.ListenAndServe()

	wsServer := websocket.NewServer(httpSrv, "/ws/pool")
	wsServer.AddEventHandler(&loadBalancerHandler{})
	wsServer.Start()

	log.Println("Connection pool test server started at ws://0.0.0.0:8080/ws/pool")
	log.Println("Waiting for client connections...")

	select {}
}
