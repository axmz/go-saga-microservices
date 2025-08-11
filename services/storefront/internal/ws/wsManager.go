package ws

import (
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSManager struct {
	mu              sync.RWMutex
	clients         map[string]map[*websocket.Conn]bool
	lastKnownStatus map[string]string
}

func NewWSManager() *WSManager {
	return &WSManager{
		clients:         make(map[string]map[*websocket.Conn]bool),
		lastKnownStatus: make(map[string]string),
	}
}

func (m *WSManager) Register(orderID string, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.clients[orderID] == nil {
		m.clients[orderID] = make(map[*websocket.Conn]bool)
	}
	m.clients[orderID][conn] = true

	if status, ok := m.lastKnownStatus[orderID]; ok {
		conn.WriteJSON(map[string]string{
			"orderId": orderID,
			"status":  status,
		})
	}
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
	m.mu.Lock()
	defer m.mu.Unlock()

	// in case broadcast is called before any clients are registered
	if m.lastKnownStatus == nil {
		m.lastKnownStatus = make(map[string]string)
	}
	m.lastKnownStatus[orderID] = status

	time.Sleep(time.Second * 3)
	for conn := range m.clients[orderID] {
		err := conn.WriteJSON(map[string]string{
			"orderId": orderID,
			"status":  status,
		})
		if err != nil {
			slog.Warn("WS write error:", "err", err)
		}
		conn.Close()
	}
}
