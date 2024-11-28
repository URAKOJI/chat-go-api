package repository

import (
	"chat-go-api/internal/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatRepository struct {
	db *mongo.Database
}

func NewChatRepository(db *mongo.Database) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateChatRoom(room *models.ChatRoom) error {
	result, err := r.db.Collection("chat_rooms").InsertOne(context.TODO(), room)
	if err != nil {
		return err
	}

	room.ID = result.InsertedID.(primitive.ObjectID)
	return err
}

// GetChatRoomsByUserID 유저가 참여한 채팅방 목록 조회
func (r *ChatRepository) GetChatRoomsByUserID(userID primitive.ObjectID) ([]models.ChatRoom, error) {
	var chatRooms []models.ChatRoom

	cursor, err := r.db.Collection("chat_rooms").Find(
		context.TODO(),
		bson.M{"members": userID}, // members 배열에 userID가 포함된 채팅방 검색
		options.Find().SetSort(bson.M{"created_at": -1}), // 최신순 정렬
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &chatRooms); err != nil {
		return nil, err
	}

	return chatRooms, nil
}
