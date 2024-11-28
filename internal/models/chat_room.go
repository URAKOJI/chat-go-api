package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type ChatRoom struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	Name      string               `bson:"name"`
	Members   []primitive.ObjectID `bson:"members"`
	CreatedAt int64                `bson:"created_at"`
}
