package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Manager 관리 구조체
type Manager struct {
	clients    map[*websocket.Conn]bool // 활성 클라이언트
	broadcast  chan []byte              // 브로드캐스트 채널
	register   chan *websocket.Conn     // 새 연결
	unregister chan *websocket.Conn     // 연결 해제
	mu         sync.Mutex               // 동시성 제어
}

// NewManager 매니저 초기화
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Run 매니저 실행
func (m *Manager) Run() {
	for {
		select {
		case conn := <-m.register:
			m.mu.Lock()
			m.clients[conn] = true
			m.mu.Unlock()
			log.Println("New client connected")

		case conn := <-m.unregister:
			m.mu.Lock()
			if _, ok := m.clients[conn]; ok {
				delete(m.clients, conn)
				conn.Close()
			}
			m.mu.Unlock()
			log.Println("Client disconnected")

		case message := <-m.broadcast:
			m.mu.Lock()
			for conn := range m.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					conn.Close()
					delete(m.clients, conn)
				}
			}
			m.mu.Unlock()
		}
	}
}

// Broadcast 메시지 브로드캐스트
func (m *Manager) Broadcast(message []byte) {
	m.broadcast <- message
}
