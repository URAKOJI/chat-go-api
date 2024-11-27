package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	Email           string             `bson:"email"`
	Password        string             `bson:"password"` // 비밀번호 해시값
	Name            string             `bson:"name"`
	Role            string             `bson:"role"` // 예: "user", "admin"
	IsEmailVerified bool               `bson:"is_email_verified"`
	CreatedAt       int64              `bson:"created_at"` // UNIX 타임스탬프
	UpdatedAt       int64              `bson:"updated_at"`
}

type LoginHistory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       primitive.ObjectID `bson:"user_id"`
	AccessToken  string             `bson:"access_token"`
	RefreshToken string             `bson:"refresh_token"`
	CreatedAt    int64              `bson:"created_at"` // UNIX 타임스탬프
	UpdatedAt    int64              `bson:"updated_at"`
}

type EmailVerification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Token     string             `bson:"token"`
	CreatedAt int64              `bson:"created_at"`
	ExpiresAt int64              `bson:"expires_at"`
}
