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
	Id        int            `json:"id,omitempty"`
	Type      string         `json:"type"`
	Content   string         `json:"content,omitempty"`
	Keys      map[int]string `json:"keys,omitempty"`
	ChatId    string         `json:"chat_id"`
	UserId    int            `json:"user_id"`
	Status    string         `json:"status,omitempty"`
	Timestamp time.Time      `json:"timestamp,omitempty"`
	MessageId int            `json:"message_id,omitempty"`
}

type BroadcastMessage struct {
	Message Message
	Keys    map[int]string
}

type Manager struct {
	Clients     map[*Client]bool
	Broadcast   chan BroadcastMessage
	Register    chan *Client
	Undergister chan *Client
	Mutex       sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		Clients:     make(map[*Client]bool),
		Broadcast:   make(chan BroadcastMessage),
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

		case broadcastMsg := <-m.Broadcast:
			message := broadcastMsg.Message
			keys := broadcastMsg.Keys

			m.Mutex.Lock()
			for client := range m.Clients {
				if client.ChatId == message.ChatId {

					clientMessage := message

					if encKey, ok := keys[client.UserId]; ok {
						clientMessage.Keys = map[int]string{client.UserId: encKey}
					} else {
						if keys != nil {
							continue
						}
					}

					if client.UserId == message.UserId {
						clientMessage.Status = "read"
					} else {
						clientMessage.Status = "delivered"
					}

					select {
					case client.Send <- m.MarshalMessage(clientMessage):
					default:
						close(client.Send)
						delete(m.Clients, client)
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

func (m *Manager) SendMessage(message Message, keys map[int]string) {
	m.Broadcast <- BroadcastMessage{Message: message, Keys: keys}
}
