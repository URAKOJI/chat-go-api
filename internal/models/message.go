package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	RoomID    primitive.ObjectID `bson:"room_id"`
	SenderID  primitive.ObjectID `bson:"sender_id"`
	Content   string             `bson:"content"`
	CreatedAt int64              `bson:"created_at"`
}
