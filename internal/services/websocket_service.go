package services

import (
	"chat-go-api/internal/models"
	"chat-go-api/internal/repository"
	"chat-go-api/internal/utils"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WebSocketManager interface {
	BroadcastToRoom(roomID string, message *models.MessageDTO) error
}

type WebSocketService struct {
	manager      WebSocketManager
	messageRepo  *repository.MessageRepository
	chatRoomRepo *repository.ChatRepository
}

func NewWebSocketService(
	manager WebSocketManager,
	messageRepo *repository.MessageRepository,
	chatRoomRepo *repository.ChatRepository,
) *WebSocketService {
	return &WebSocketService{
		manager:      manager,
		messageRepo:  messageRepo,
		chatRoomRepo: chatRoomRepo,
	}
}

func (s *WebSocketService) GetUserName(userID primitive.ObjectID) (string, error) {
	user, err := s.messageRepo.GetUserByID(userID)
	if err != nil {
		return "", err
	}
	return user.Name, nil
}

func (s *WebSocketService) HandleIncomingMessage(roomID, senderID string, data []byte) error {
	var msg struct {
		Content string `json:"content"`
	}

	// 메시지 파싱
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	// 메시지 모델 생성
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return err
	}

	senderObjectID, err := primitive.ObjectIDFromHex(senderID)
	if err != nil {
		return err
	}

	message := &models.Message{
		RoomID:    roomObjectID,
		SenderID:  senderObjectID,
		Content:   msg.Content,
		CreatedAt: time.Now().Unix(),
	}

	// 메시지 저장
	if err := s.messageRepo.SaveMessage(message); err != nil {
		return err
	}

	// 메시지 DTO 생성
	messageDTO, err := utils.ToMessageDTO(message, s.GetUserName)
	if err != nil {
		return err
	}

	// 브로드캐스트
	return s.manager.BroadcastToRoom(roomID, messageDTO)
}
