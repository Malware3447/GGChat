package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	//"github.com/sirupsen/logrus"
)

type Client struct {
	Id     string
	UserId int
	ChatId string
	Conn   *websocket.Conn
	Send   chan []byte
}

type Message struct {
	Id        int       `json:"id,omitempty"` // ID самого сообщения
	Type      string    `json:"type"`
	Content   string    `json:"content,omitempty"`
	ChatId    string    `json:"chat_id"`
	UserId    int       `json:"user_id"`
	Status    string    `json:"status,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	MessageId int       `json:"message_id,omitempty"` // ID сообщения (для 'read_receipt')
}

type Manager struct {
	Clients     map[*Client]bool
	Broadcast   chan Message
	Register    chan *Client
	Undergister chan *Client
	Mutex       sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		Clients:     make(map[*Client]bool),
		Broadcast:   make(chan Message),
		Register:    make(chan *Client),
		Undergister: make(chan *Client),
	}
}

func (m *Manager) Run() {
	for {
		select {
		case client := <-m.Register:
			m.Mutex.Lock()
			m.Clients[client] = true
			m.Mutex.Unlock()

		case client := <-m.Undergister:
			m.Mutex.Lock()
			if _, ok := m.Clients[client]; ok {
				delete(m.Clients, client)
				close(client.Send)
				client.Conn.Close()
			}
			m.Mutex.Unlock()

		case message := <-m.Broadcast:
			m.Mutex.Lock()
			for Client := range m.Clients {
				if Client.ChatId == message.ChatId {
					select {
					case Client.Send <- m.MarshalMessage(message):
					default:
						close(Client.Send)
						delete(m.Clients, Client)
					}
				}
			}
			m.Mutex.Unlock()
		}

	}
}

func (m *Manager) MarshalMessage(msg Message) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		return []byte{}
	}
	return data
}

func (m *Manager) SendMessage(message Message) {
	m.Broadcast <- message
}
