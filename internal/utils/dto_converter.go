package utils

import (
	"chat-go-api/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToMessageDTO 메시지를 DTO로 변환
func ToMessageDTO(message *models.Message, getUserName func(userID primitive.ObjectID) (string, error)) (*models.MessageDTO, error) {
	// 작성자 이름 조회
	senderName, err := getUserName(message.SenderID)
	if err != nil {
		return nil, err
	}

	return &models.MessageDTO{
		ID:         message.ID,
		RoomID:     message.RoomID,
		SenderID:   message.SenderID,
		SenderName: senderName,
		Content:    message.Content,
		CreatedAt:  message.CreatedAt,
	}, nil
}
