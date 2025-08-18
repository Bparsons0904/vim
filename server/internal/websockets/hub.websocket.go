package websockets

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	STATUS_UNAUTHENTICATED = iota
	STATUS_PENDING
	STATUS_AUTHENTICATED
	STATUS_CLOSED
)

type Hub struct {
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	clients    map[string]*Client
	mutex      sync.RWMutex
}

func (h *Hub) run(m *Manager) {
	for {
		select {
		case client := <-h.register:
			m.registerClient(client)

		case client := <-h.unregister:
			func() {
				defer func() {
					if r := recover(); r != nil {
						_ = r // Explicitly ignore recovered value
					}
				}()
				close(client.send)
			}()
			m.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message, m)
		}
	}
}

func (m *Manager) unregisterClient(client *Client) {
	log := m.log.Function("unregisterClient")
	log.Info(
		"Unregistering client",
		"clientID",
		client.ID,
		"userID",
		client.UserID,
		"status",
		client.Status,
	)

	m.hub.mutex.Lock()
	defer m.hub.mutex.Unlock()

	delete(m.hub.clients, client.ID)

	log.Info(
		"Client unregistered and removed from local storage",
		"clientID",
		client.ID,
		"userID",
		client.UserID,
	)
}

func (m *Manager) registerClient(client *Client) {
	log := m.log.Function("registerClient")
	log.Info("Registering client", "clientID", client.ID, "status", client.Status)

	m.hub.mutex.Lock()
	defer m.hub.mutex.Unlock()

	m.hub.clients[client.ID] = client

	log.Info(
		"Client registered",
		"clientID",
		client.ID,
		"userID",
		client.UserID,
		"status",
		client.Status,
	)
}

func (h *Hub) broadcastMessage(message Message, m *Manager) {
	log := m.log.Function("broadcastMessage")

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if len(h.clients) == 0 {
		log.Info("No active clients to broadcast to", "messageID", message.ID)
		return
	}

	sentCount := 0
	totalClients := len(h.clients)

	for clientID, client := range h.clients {
		if client.Status != STATUS_AUTHENTICATED {
			continue
		}

		select {
		case client.send <- message:
			sentCount++
		default:
			go func(c *Client, cID string, msg Message) {
				select {
				case c.send <- msg:
					log.Info("Message sent after retry", "clientID", cID)
				case <-time.After(5 * time.Second):
					_ = log.Error("Client too slow, disconnecting", "clientID", cID)
					m.hub.unregister <- c
				}
			}(client, clientID, message)
		}
	}

	log.Info(
		"Broadcast complete",
		"messageID",
		message.ID,
		"sentTo",
		sentCount,
		"totalClients",
		totalClients,
	)
}

func (m *Manager) promoteClientToAuthenticated(client *Client) {
	log := m.log.Function("promoteClientToAuthenticated")

	if client.Status != STATUS_AUTHENTICATED {
		log.Warn("Attempted to promote non-authenticated client", "clientID", client.ID)
		return
	}

	log.Info(
		"Client promoted to authenticated",
		"clientID",
		client.ID,
		"userID",
		client.UserID,
	)
}

func (m *Manager) SendMessageToUser(userID uuid.UUID, message Message) {
	log := m.log.Function("SendMessageToUser")

	m.hub.mutex.RLock()
	defer m.hub.mutex.RUnlock()

	sentCount := 0
	totalUserConnections := 0

	for clientID, client := range m.hub.clients {
		if client.Status == STATUS_AUTHENTICATED && client.UserID == userID {
			totalUserConnections++
			select {
			case client.send <- message:
				sentCount++
			default:
				go func(c *Client, cID string, msg Message, uID uuid.UUID) {
					select {
					case c.send <- msg:
						log.Info("Message sent after retry", "clientID", cID, "userID", uID)
					case <-time.After(5 * time.Second):
						_ = log.Error(
							"Client too slow, disconnecting",
							"clientID",
							cID,
							"userID",
							uID,
						)
						m.hub.unregister <- c
					}
				}(client, clientID, message, userID)
			}
		}
	}

	if totalUserConnections == 0 {
		log.Info("No connections found for user", "userID", userID)
		return
	}

	log.Info(
		"Message sent to user connections",
		"userID",
		userID,
		"messageID",
		message.ID,
		"sentTo",
		sentCount,
		"totalConnections",
		totalUserConnections,
	)
}
