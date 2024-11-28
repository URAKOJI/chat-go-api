package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type MessageDTO struct {
	ID         primitive.ObjectID `json:"id"`
	RoomID     primitive.ObjectID `json:"room_id"`
	SenderID   primitive.ObjectID `json:"sender_id"`
	SenderName string             `json:"sender_name"` // 작성자 이름
	Content    string             `json:"content"`
	CreatedAt  int64              `json:"created_at"`
}
