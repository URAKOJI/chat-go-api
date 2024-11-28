package common

import (
	"chat-go-api/internal/models"

	"github.com/gorilla/websocket"
)

// BroadcastMessage 브로드캐스트 메시지 구조체
type BroadcastMessage struct {
	RoomID  string
	Message *models.Message
}

// RegisterMessage 연결 등록 구조체
type RegisterMessage struct {
	RoomID string
	UserID string
	Conn   *websocket.Conn
}

// UnregisterMessage 연결 해제 구조체
type UnregisterMessage struct {
	RoomID string
	Conn   *websocket.Conn
}
