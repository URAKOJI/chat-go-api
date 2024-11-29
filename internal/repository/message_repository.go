package repository

import (
	"chat-go-api/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MessageRepository struct {
	db *mongo.Database
}

func NewMessageRepository(db *mongo.Database) *MessageRepository {
	return &MessageRepository{db: db}
}

func (r *MessageRepository) SaveMessage(msg *models.Message) error {
	result, err := r.db.Collection("messages").InsertOne(context.TODO(), msg)

	if err != nil {
		return err
	}

	msg.ID = result.InsertedID.(primitive.ObjectID)
	return err
}

// GetMessagesByRoomID 특정 채팅방의 메시지 로드
func (r *MessageRepository) GetMessagesByRoomID(roomID primitive.ObjectID) ([]models.Message, error) {
	var messages []models.Message

	// MongoDB 쿼리 실행
	cursor, err := r.db.Collection("messages").Find(
		context.TODO(),
		bson.M{"room_id": roomID},
		options.Find().SetSort(bson.M{"created_at": 1}), // 시간 순서로 정렬
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	// 쿼리 결과를 messages로 디코딩
	if err := cursor.All(context.TODO(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepository) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetTotalMessagesCount 특정 채팅방의 총 메시지 개수 반환
func (r *MessageRepository) GetTotalMessagesCount(roomID primitive.ObjectID) (int64, error) {
	total, err := r.db.Collection("messages").CountDocuments(
		context.TODO(),
		bson.M{"room_id": roomID}, // 특정 room_id의 메시지 개수 계산
	)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *MessageRepository) GetMessagesByRoomIDWithPagination(roomID primitive.ObjectID, limit int64, page int64) ([]*models.Message, error) {
	skip := (page - 1) * limit

	// 전체 메시지 수 확인
	totalMessages, err := r.db.Collection("messages").CountDocuments(context.TODO(), bson.M{"room_id": roomID})
	if err != nil {
		return nil, err
	}

	// 페이지 초과 확인
	if skip >= totalMessages {
		return []*models.Message{}, nil // 빈 배열 반환
	}

	queryOptions := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(skip)
	if totalMessages > limit {
		queryOptions.SetLimit(limit)
	}

	cursor, err := r.db.Collection("messages").Find(
		context.TODO(),
		bson.M{"room_id": roomID},
		queryOptions,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var messages []*models.Message
	if err := cursor.All(context.TODO(), &messages); err != nil {
		return nil, err
	}

	return messages, nil
}
