package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/yunis-du/websocket-go/websocket"
)

type multiNamespaceHandler struct{}

func (h *multiNamespaceHandler) OnMessage(c websocket.Speaker, m websocket.Message) {}

func (h *multiNamespaceHandler) OnStructMessage(c websocket.Speaker, m *websocket.StructMessage) {
	log.Printf("📨 [%s] Received message: Type=%s, Status=%d", m.Namespace, m.Type, m.Status)

	// for pretty print
	var data any
	if err := json.Unmarshal(m.Obj, &data); err == nil {
		prettyJSON, _ := json.MarshalIndent(data, "", "  ")
		fmt.Printf("Data:\n%s\n\n", string(prettyJSON))
	}
}

func (h *multiNamespaceHandler) OnConnected(c websocket.Speaker) {
	log.Println("✓ Connected to server")
}

func (h *multiNamespaceHandler) OnDisconnected(c websocket.Speaker) {
	log.Println("✗ Disconnected")
}

func main() {
	hostname, _ := os.Hostname()

	u := &url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/api"}
	opt := websocket.ClientOpts{
		Protocol:  websocket.JSONProtocol,
		QueueSize: 1000,
	}

	wsClient := websocket.NewClient(hostname, u, opt)
	structClient := wsClient.UpgradeToStructSpeaker()

	// Listen to all namespaces
	structClient.AddStructMessageHandler(&multiNamespaceHandler{}, []string{"user", "order", "notification"})
	wsClient.Start()

	time.Sleep(500 * time.Millisecond)

	// Test User namespace
	fmt.Println("=== Test User namespace ===")
	userMsg := websocket.NewStructMessage("user", "get_profile", nil)
	_, err := structClient.Request(userMsg, 3*time.Second)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
	} else {
		log.Printf("✓ Get user profile success")
	}

	time.Sleep(1 * time.Second)

	// Update user profile
	updateData := map[string]string{
		"username": "jane_doe",
		"email":    "jane@example.com",
	}
	updateMsg := websocket.NewStructMessage("user", "update_profile", updateData)
	_, err = structClient.Request(updateMsg, 3*time.Second)
	if err != nil {
		log.Printf("❌ Update failed: %v", err)
	} else {
		log.Printf("✓ Update user profile success")
	}

	time.Sleep(1 * time.Second)

	// Test Order namespace
	fmt.Println("\n=== Test Order namespace ===")
	orderMsg := websocket.NewStructMessage("order", "list_orders", nil)
	_, err = structClient.Request(orderMsg, 3*time.Second)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
	} else {
		log.Printf("✓ Get order list success")
	}

	time.Sleep(1 * time.Second)

	// Create order
	newOrder := map[string]any{
		"product":  "Keyboard",
		"quantity": 1,
		"price":    79.99,
	}
	createOrderMsg := websocket.NewStructMessage("order", "create_order", newOrder)
	_, err = structClient.Request(createOrderMsg, 3*time.Second)
	if err != nil {
		log.Printf("❌ Create order failed: %v", err)
	} else {
		log.Printf("✓ Create order success")
	}

	time.Sleep(1 * time.Second)

	// Test Notification namespace
	fmt.Println("\n=== Test Notification namespace ===")
	notifyMsg := websocket.NewStructMessage("notification", "subscribe", nil)
	_, err = structClient.Request(notifyMsg, 3*time.Second)
	if err != nil {
		log.Printf("❌ Subscribe failed: %v", err)
	} else {
		log.Printf("✓ Subscribe success, waiting for push...")
	}

	// Wait for Notification
	time.Sleep(5 * time.Second)

	fmt.Println("\nAll test completed")
	wsClient.Stop()
}
