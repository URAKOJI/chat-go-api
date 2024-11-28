package websocket

import (
	"chat-go-api/internal/common"
	"chat-go-api/internal/models"
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type ManagerInterface interface {
	BroadcastToRoom(roomID string, message models.MessageDTO) error
	RegisterClient(roomID string, conn *websocket.Conn)
	UnregisterClient(roomID string, conn *websocket.Conn)
}

type Manager struct {
	rooms      map[string]map[*websocket.Conn]string // roomID -> conn -> userID
	broadcast  chan common.BroadcastMessage
	register   chan common.RegisterMessage
	unregister chan common.UnregisterMessage
}

func NewManager() *Manager {
	return &Manager{
		rooms:      make(map[string]map[*websocket.Conn]string),
		broadcast:  make(chan common.BroadcastMessage),
		register:   make(chan common.RegisterMessage),
		unregister: make(chan common.UnregisterMessage),
	}
}

func (m *Manager) RegisterClientWithUser(roomID string, conn *websocket.Conn, userID string) {
	if _, ok := m.rooms[roomID]; !ok {
		m.rooms[roomID] = make(map[*websocket.Conn]string)
	}
	m.rooms[roomID][conn] = userID
}

func (m *Manager) GetUserID(roomID string, conn *websocket.Conn) (string, bool) {
	if room, ok := m.rooms[roomID]; ok {
		if userID, exists := room[conn]; exists {
			return userID, true
		}
	}
	return "", false
}

// Run 및 기타 기존 메서드는 동일
func (m *Manager) BroadcastToRoom(roomID string, message *models.MessageDTO) error {
	if clients, ok := m.rooms[roomID]; ok {
		for conn := range clients {
			// 메시지 직렬화
			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Failed to serialize message: %v", err)
				continue
			}

			// 메시지 전송
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Failed to send message: %v", err)
				conn.Close()
				delete(clients, conn)
			}
		}
	}
	return nil
}

func (m *Manager) RegisterClient(roomID string, conn *websocket.Conn) {
	m.register <- common.RegisterMessage{
		RoomID: roomID,
		Conn:   conn,
	}
}

func (m *Manager) UnregisterClient(roomID string, conn *websocket.Conn) {
	m.unregister <- common.UnregisterMessage{
		RoomID: roomID,
		Conn:   conn,
	}
}

func (m *Manager) Run() {
	for {
		select {
		case msg := <-m.register:
			// 클라이언트 등록
			if _, ok := m.rooms[msg.RoomID]; !ok {
				m.rooms[msg.RoomID] = make(map[*websocket.Conn]string)
			}
			m.rooms[msg.RoomID][msg.Conn] = msg.UserID // UserID 저장

		case msg := <-m.unregister:
			// 클라이언트 해제
			if clients, ok := m.rooms[msg.RoomID]; ok {
				if _, exists := clients[msg.Conn]; exists {
					delete(clients, msg.Conn)
					msg.Conn.Close()
				}
				if len(clients) == 0 {
					delete(m.rooms, msg.RoomID)
				}
			}

		case msg := <-m.broadcast:
			// 메시지 브로드캐스트
			if clients, ok := m.rooms[msg.RoomID]; ok {
				for conn := range clients {
					data, err := json.Marshal(msg.Message)
					if err != nil {
						log.Printf("Failed to serialize message: %v", err)
						continue
					}
					if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
						conn.Close()
						delete(clients, conn)
					}
				}
			}
		}
	}
}
