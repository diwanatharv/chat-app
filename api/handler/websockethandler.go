package handler

import (
	"chat-app/pkg/enums"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	//This is the size of the buffer that holds incoming data from the client
	// If a client sends a message that's 1.5 KB in size, it will fit into the 2 KB buffer, and the server will read it in one go. But if the message is 3 KB, it will be split, and the server will need to read it in two steps (first 2 KB, then 1 KB).
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	//This is a security check to determine whether to accept WebSocket connections from any origin (i.e., any domain). In this case, it always returns true, meaning it accepts connections from any origin.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

type Connection struct {
	Conn *websocket.Conn
	ID   string
}

var (
	customerConnections = make(map[string]*Connection) // Map for customer connections
	agentConnections    = make(map[string]*Connection) // Map for agent connections
	mu                  sync.Mutex                     // Mutex for handling concurrent map access
)

func addCustomerConnection(customerID string, conn *websocket.Conn) {
	mu.Lock()
	customerConnections[customerID] = &Connection{Conn: conn, ID: customerID}
	mu.Unlock()
}

func removeCustomerConnection(customerID string) {
	mu.Lock()
	delete(customerConnections, customerID)
	mu.Unlock()
}

func addAgentConnection(agentID string, conn *websocket.Conn) {
	mu.Lock()
	agentConnections[agentID] = &Connection{Conn: conn, ID: agentID}
	mu.Unlock()
}

func removeAgentConnection(agentID string) {
	mu.Lock()
	delete(agentConnections, agentID)
	mu.Unlock()
}
func upgradeConnection(e echo.Context) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(e.Response(), e.Request(), nil)
	if err != nil {
		logrus.WithError(err).Error("Unable to upgrade websocket connection")
		return nil, e.JSON(http.StatusInternalServerError, enums.InternalServeerror)
	}
	return conn, nil
}
func CustomerWebSocketHandler(c echo.Context) error {
	// Generate a unique customer ID
	customerID := uuid.NewString()

	conn, err := upgradeConnection(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error upgrading connection")
	}

	// Handle customer connection in a separate goroutine
	go handleCustomerConnection(customerID, conn)

	return nil
}

// Handle customer WebSocket connection
func handleCustomerConnection(customerID string, conn *websocket.Conn) {
	defer conn.Close() // Ensure the connection is closed when the function returns

	addCustomerConnection(customerID, conn)

	logrus.Infof("Customer connected: %s", customerID)

	for {
		// Read messages from the customer
		_, message, err := conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("Error reading customer message, disconnecting")
			break
		}
		logrus.Infof("Received message from customer %s: %s", customerID, message)

		// TODO: Broadcast message to agent (this will be handled in future tasks)
	}

	// Handle disconnection
	removeCustomerConnection(customerID)

	logrus.Infof("Customer disconnected: %s", customerID)
}

// WebSocket handler for agents
func AgentWebSocketHandler(c echo.Context) error {
	agentID := uuid.NewString()

	conn, err := upgradeConnection(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Error upgrading connection")
	}

	// Handle agent connection in a separate goroutine
	go handleAgentConnection(agentID, conn)

	return nil
}

// Handle agent WebSocket connection
func handleAgentConnection(agentID string, conn *websocket.Conn) {
	defer conn.Close() // Ensure the connection is closed when the function returns

	// Add agent connection to the map
	addAgentConnection(agentID, conn)

	logrus.Infof("Agent connected: %s", agentID)

	for {
		// Read messages from the agent
		_, message, err := conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("Error reading agent message, disconnecting")
			break
		}
		logrus.Infof("Received message from agent %s: %s", agentID, message)

		// TODO: Broadcast message to customer (this will be handled in future tasks)
	}

	// Handle disconnection
	removeAgentConnection(agentID)

	logrus.Infof("Agent disconnected: %s", agentID)
}
