package websocket

import (
	"chat-go-api/internal/services"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(manager *Manager, wsService *services.WebSocketService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// WebSocket 연결 업그레이드
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
			return
		}

		// 토큰 인증
		tokenString := r.URL.Query().Get("token")
		if tokenString == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("Missing token"))
			conn.Close()
			return
		}

		// 토큰 검증
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("access-secret-key"), nil
		})
		if err != nil || !token.Valid {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid token"))
			conn.Close()
			return
		}

		// 사용자 ID 추출
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid token claims"))
			conn.Close()
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			conn.WriteMessage(websocket.TextMessage, []byte("Invalid user_id in token"))
			conn.Close()
			return
		}

		// 채팅방 ID 가져오기
		roomID := r.URL.Query().Get("room_id")
		if roomID == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("Missing room_id"))
			conn.Close()
			return
		}

		// 클라이언트 등록
		manager.RegisterClientWithUser(roomID, conn, userID)

		// 메시지 수신 루프
		go func() {
			defer manager.UnregisterClient(roomID, conn)
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					break
				}

				// SenderID 가져오기
				senderID, ok := manager.GetUserID(roomID, conn)
				if !ok {
					log.Println("Failed to get senderID")
					continue
				}

				// 메시지 처리
				if err := wsService.HandleIncomingMessage(roomID, senderID, message); err != nil {
					log.Printf("Failed to handle WebSocket message: %v", err)
				}
			}
		}()
	}
}
