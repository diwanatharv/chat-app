package handler

import (
	cache "chat-app/pkg/redis"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// the function always returns true, which means it allows connections from any origin.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketServer struct {
	clients     map[*websocket.Conn]bool
	redisClient cache.RedisClient
	mu          sync.Mutex // Protects the clients map
}

func NewWebSocketServer(redisClient cache.RedisClient) *WebSocketServer {
	return &WebSocketServer{
		clients:     make(map[*websocket.Conn]bool),
		redisClient: redisClient,
	}
}

// The HandleConnections method upgrades an HTTP connection to a WebSocket connection, stores the connection in a map of active clients, continuously reads messages from the connection, and publishes those messages to a Redis channel. If any errors occur during these processes, appropriate error messages are printed, and the connection is handled accordingly.
func (s *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer func() {
		conn.Close()
		s.mu.Lock()
		_, exists := s.clients[conn]
		if exists {
			delete(s.clients, conn)
		}
		s.mu.Unlock()
	}()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error while reading message:", err)
			break
		}

		err = s.redisClient.PublishMessage("chat_channel", string(msg))
		if err != nil {
			fmt.Println("Error while publishing message to Redis:", err)
		}
	}
}

// method is designed to send a given message to all active WebSocket clients. It iterates over the clients map, sends the message to each client, and handles any errors that occur during the process by closing and removing the problematic connection from the map. This ensures that only active and functioning connections remain in the clients map.
func (s *WebSocketServer) BroadcastMessage(msg []byte) {
	for conn := range s.clients {
		err := conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			fmt.Println("Error while writing message:", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// Add this method to start Redis subscription and handle incoming messages
// method sets up a subscription to a Redis channel and ensures that any messages received on that channel are broadcast to all connected WebSocket clients. This allows for real-time message distribution from Redis to WebSocket clients.
func (s *WebSocketServer) StartRedisSubscription() {
	messageChannel := make(chan string)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		err := s.redisClient.SubscribeToChannel(ctx, "chat_channel", messageChannel)
		if err != nil {
			fmt.Println("Error while subscribing to Redis channel:", err)
		}
	}()

	for msg := range messageChannel {
		s.BroadcastMessage([]byte(msg))
	}
}
