package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	httpServer "github.com/yunis-du/websocket-go/http"
	"github.com/yunis-du/websocket-go/websocket"
)

// UserService Handle user-related messages
type UserService struct{}

func (s *UserService) OnMessage(c websocket.Speaker, m websocket.Message) {}

func (s *UserService) OnStructMessage(c websocket.Speaker, m *websocket.StructMessage) {
	log.Printf("[UserService] Received message: Type=%s", m.Type)

	switch m.Type {
	case "get_profile":
		profile := map[string]any{
			"id":       12345,
			"username": "john_doe",
			"email":    "john@example.com",
			"created":  time.Now().Format(time.RFC3339),
		}
		reply := m.Reply(profile, "profile", http.StatusOK)
		_ = c.SendMessage(reply)

	case "update_profile":
		var data map[string]any
		json.Unmarshal(m.Obj, &data)
		log.Printf("Update user profile: %v", data)

		reply := m.Reply(map[string]string{"status": "success"}, "update_result", http.StatusOK)
		_ = c.SendMessage(reply)
	}
}

func (s *UserService) OnConnected(c websocket.Speaker) {
	log.Println("[UserService] Client connected")
}

func (s *UserService) OnDisconnected(c websocket.Speaker) {
	log.Println("[UserService] Client disconnected")
}

// OrderService Handle order-related messages
type OrderService struct{}

func (s *OrderService) OnMessage(c websocket.Speaker, m websocket.Message) {}

func (s *OrderService) OnStructMessage(c websocket.Speaker, m *websocket.StructMessage) {
	log.Printf("[OrderService] Received message: Type=%s", m.Type)

	switch m.Type {
	case "list_orders":
		orders := []map[string]any{
			{"id": 1001, "product": "Laptop", "price": 999.99, "status": "shipped"},
			{"id": 1002, "product": "Mouse", "price": 29.99, "status": "delivered"},
		}
		reply := m.Reply(orders, "orders", http.StatusOK)
		_ = c.SendMessage(reply)

	case "create_order":
		var orderData map[string]any
		json.Unmarshal(m.Obj, &orderData)
		log.Printf("Create order: %v", orderData)

		newOrder := map[string]any{
			"id":      1003,
			"status":  "pending",
			"created": time.Now().Format(time.RFC3339),
		}
		reply := m.Reply(newOrder, "order_created", http.StatusCreated)
		_ = c.SendMessage(reply)
	}
}

func (s *OrderService) OnConnected(c websocket.Speaker) {
	log.Println("[OrderService] Client connected")
}

func (s *OrderService) OnDisconnected(c websocket.Speaker) {
	log.Println("[OrderService] Client disconnected")
}

// NotificationService Handle notification-related messages
type NotificationService struct{}

func (s *NotificationService) OnMessage(c websocket.Speaker, m websocket.Message) {}

func (s *NotificationService) OnStructMessage(c websocket.Speaker, m *websocket.StructMessage) {
	log.Printf("[NotificationService] Received message: Type=%s", m.Type)

	switch m.Type {
	case "subscribe":
		log.Println("Client subscribe notifications")
		reply := m.Reply(map[string]string{"status": "subscribed"}, "subscribe_result", http.StatusOK)
		_ = c.SendMessage(reply)

		// Simulate sending notification
		go func() {
			time.Sleep(2 * time.Second)
			notification := websocket.NewStructMessage("notification", "new_message", map[string]string{
				"title": "New message",
				"body":  "You have a new message",
			})
			_ = c.SendMessage(notification)
		}()
	}
}

func (s *NotificationService) OnConnected(c websocket.Speaker) {
	log.Println("[NotificationService] Client connected")
}

func (s *NotificationService) OnDisconnected(c websocket.Speaker) {
	log.Println("[NotificationService] Client disconnected")
}

func main() {
	hostname, _ := os.Hostname()
	httpSrv := httpServer.NewServer(hostname, "0.0.0.0", 8080)
	httpSrv.ListenAndServe()

	wsServer := websocket.NewServer(httpSrv, "/ws/api")
	structServer := websocket.NewStructServer(wsServer)

	// Register handlers for different namespaces
	structServer.AddStructMessageHandler(&UserService{}, []string{"user"})
	structServer.AddStructMessageHandler(&OrderService{}, []string{"order"})
	structServer.AddStructMessageHandler(&NotificationService{}, []string{"notification"})

	structServer.Start()

	log.Println("Multi-namespace server started at ws://0.0.0.0:8080/ws/api")
	log.Println("Supported namespaces: user, order, notification")

	select {}
}
