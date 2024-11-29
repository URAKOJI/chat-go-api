package services

import (
	"chat-go-api/internal/common"
	"chat-go-api/internal/models"
	"chat-go-api/internal/repository"
	"chat-go-api/internal/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService struct {
	chatRepo    *repository.ChatRepository
	messageRepo *repository.MessageRepository
}

func NewChatService(chatRepo *repository.ChatRepository, messageRepo *repository.MessageRepository) *ChatService {
	return &ChatService{chatRepo: chatRepo, messageRepo: messageRepo}
}

func (s *ChatService) CreateChatRoom(name string, memberIDs []primitive.ObjectID) (*models.ChatRoom, error) {
	room := &models.ChatRoom{
		Name:      name,
		Members:   memberIDs,
		CreatedAt: time.Now().Unix(),
	}
	err := s.chatRepo.CreateChatRoom(room)
	return room, err
}

func (s *ChatService) SaveMessage(msg *models.Message) error {
	return s.messageRepo.SaveMessage(msg)
}

// GetChatRoomMessages 특정 채팅방의 메시지 로드
func (s *ChatService) GetChatRoomMessages(roomID primitive.ObjectID) ([]models.Message, error) {
	return s.messageRepo.GetMessagesByRoomID(roomID)
}

// GetUserChatRooms 유저가 참여한 채팅방 목록 조회
func (s *ChatService) GetUserChatRooms(userID string) ([]models.ChatRoom, error) {
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	return s.chatRepo.GetChatRoomsByUserID(uid)
}

func (s *ChatService) GetUserName(userID primitive.ObjectID) (string, error) {
	user, err := s.messageRepo.GetUserByID(userID)
	if err != nil {
		return common.UNKNOWN_USER_NAME, nil
	}
	return user.Name, nil
}

// GetTotalMessagesCount 특정 채팅방의 메시지 총 개수
func (s *ChatService) GetTotalMessagesCount(roomID string) (int64, error) {
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return 0, err
	}
	return s.messageRepo.GetTotalMessagesCount(roomObjectID)
}

func (s *ChatService) GetRecentMessages(roomID string, limit int64, page int64) ([]*models.MessageDTO, error) {
	// RoomID를 ObjectID로 변환
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, err
	}

	// 메시지 조회
	messages, err := s.messageRepo.GetMessagesByRoomIDWithPagination(roomObjectID, limit, page)
	if err != nil {
		return nil, err
	}

	// 메시지를 DTO로 변환
	var messageDTOs []*models.MessageDTO
	for _, message := range messages {
		dto, err := utils.ToMessageDTO(message, s.GetUserName)
		if err != nil {
			return nil, err
		}
		messageDTOs = append(messageDTOs, dto)
	}

	// 최신 순으로 정렬 (MongoDB의 결과는 최신순이므로 reverse 필요)
	for i, j := 0, len(messageDTOs)-1; i < j; i, j = i+1, j-1 {
		messageDTOs[i], messageDTOs[j] = messageDTOs[j], messageDTOs[i]
	}

	return messageDTOs, nil
}
