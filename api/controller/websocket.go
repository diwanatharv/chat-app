package controller

import (
	cache "chat-app/pkg/redis"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Client struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*Client]bool) // Active WebSocket clients
var clientsMutex sync.Mutex          // Mutex to protect the clients map

// HandleWebSocketConnection handles new WebSocket connections
func HandleWebSocketConnection(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return err
	}
	defer ws.Close()

	client := &Client{conn: ws}

	// Add the new client to the clients map
	clientsMutex.Lock()
	clients[client] = true
	clientsMutex.Unlock()

	// Clean up when the client disconnects
	defer func() {
		clientsMutex.Lock()
		delete(clients, client)
		clientsMutex.Unlock()
	}()

	// Listen for messages from the WebSocket client and publish them to Redis
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return err
		}

		// Publish message to Redis channel
		err = cache.PublishMessage("chat_room", string(msg))
		if err != nil {
			return err
		}
	}
}

// broadcastToClients sends a message to all connected WebSocket clients
func broadcastToClients(message string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		client.mutex.Lock() // Lock the client to avoid concurrent writes
		err := client.conn.WriteMessage(websocket.TextMessage, []byte(message))
		client.mutex.Unlock() // Unlock after writing

		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			client.conn.Close()
			clientsMutex.Lock()
			delete(clients, client)
			clientsMutex.Unlock()
		}
	}
}

// StartRedisSubscription subscribes to the Redis channel once at startup
func StartRedisSubscription() {
	cache.SubscribeToChannel("chat_room", func(message string) {
		// When a message is received from Redis, broadcast it to the WebSocket clients
		broadcastToClients(message)
	})
}
