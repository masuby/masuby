package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"inform-system/internal/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a WebSocket client connection
type Client struct {
	ID       string
	UserID   int64
	UserRole string
	ADM1Code string // For regional committee filtering
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
}

// Hub manages WebSocket connections and message broadcasting
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client connected: %s (User: %d)", client.ID, client.UserID)

			// Send connection confirmation
			msg := models.WSMessage{
				Type:      "connection",
				Action:    "connected",
				Data:      map[string]interface{}{"client_id": client.ID},
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(msg); err == nil {
				client.Send <- data
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				log.Printf("Client disconnected: %s", client.ID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastDataEntry sends a data entry update to all connected clients
func (h *Hub) BroadcastDataEntry(entry *models.DataEntry, action string) {
	msg := models.WSMessage{
		Type:      "data_entry",
		Action:    action,
		Data:      entry,
		UserID:    entry.EnteredByID,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling data entry message: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastRiskUpdate sends a risk score update to all connected clients
func (h *Hub) BroadcastRiskUpdate(score *models.RiskScore, action string) {
	msg := models.WSMessage{
		Type:      "risk_update",
		Action:    action,
		Data:      score,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling risk update message: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastNotification sends a notification to all connected clients
func (h *Hub) BroadcastNotification(title, message string, level string) {
	msg := models.WSMessage{
		Type:   "notification",
		Action: level, // info, warning, success, error
		Data: map[string]interface{}{
			"title":   title,
			"message": message,
		},
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling notification message: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastMessage sends a generic message to all connected clients
func (h *Hub) BroadcastMessage(msgType, action string, payload interface{}) {
	msg := models.WSMessage{
		Type:      msgType,
		Action:    action,
		Data:      payload,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.broadcast <- data
}

// BroadcastToRegion sends a message only to clients from a specific region
func (h *Hub) BroadcastToRegion(adm1Code string, msg models.WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling regional message: %v", err)
		return
	}

	h.mutex.RLock()
	for client := range h.clients {
		if client.ADM1Code == adm1Code || client.UserRole == "admin" {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// GetConnectedClients returns the number of connected clients
func (h *Hub) GetConnectedClients() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// GetClientsByRegion returns clients grouped by region
func (h *Hub) GetClientsByRegion() map[string]int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	regions := make(map[string]int)
	for client := range h.clients {
		if client.ADM1Code != "" {
			regions[client.ADM1Code]++
		} else {
			regions["national"]++
		}
	}
	return regions
}

// HandleWebSocket handles WebSocket upgrade and client management
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID int64, userRole, adm1Code string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		ID:       generateClientID(),
		UserID:   userID,
		UserRole: userRole,
		ADM1Code: adm1Code,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024) // 512KB max message size
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process incoming messages if needed
		var msg models.WSMessage
		if err := json.Unmarshal(message, &msg); err == nil {
			// Handle client messages (e.g., acknowledgments, requests)
			log.Printf("Received message from client %s: %s", c.ID, msg.Type)
		}
	}
}

// generateClientID creates a unique client identifier
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
