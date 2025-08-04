package ws

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

type WSManager struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // orderID => connections
}

func NewWSManager() *WSManager {
	return &WSManager{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

func (m *WSManager) Register(orderID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.clients[orderID] == nil {
		m.clients[orderID] = make(map[*websocket.Conn]bool)
	}
	m.clients[orderID][conn] = true
}

func (m *WSManager) Unregister(orderID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if clients, ok := m.clients[orderID]; ok {
		delete(clients, conn)
		if len(clients) == 0 {
			delete(m.clients, orderID)
		}
	}
}

func (m *WSManager) Broadcast(orderID string, status string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for conn := range m.clients[orderID] {
		err := conn.WriteJSON(map[string]string{
			"orderId": orderID,
			"status":  status,
		})
		if err != nil {
			slog.Warn("WS write error:", "err", err)
		}
	}
}
