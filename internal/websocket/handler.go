package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 모든 요청 허용 (보안 필요 시 조건 추가)
		return true
	},
}

// WebSocketHandler 웹소켓 핸들러
func WebSocketHandler(manager *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		manager.register <- conn

		// 메시지 수신 루프
		go func() {
			defer func() {
				manager.unregister <- conn
			}()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					break
				}
				manager.Broadcast(message) // 받은 메시지를 브로드캐스트
			}
		}()
	}
}
