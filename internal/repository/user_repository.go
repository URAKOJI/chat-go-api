package repository

import (
	"chat-go-api/internal/models"
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	db *mongo.Database
}

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{db: db}
}

// 사용자 조회
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 로그인 이력 추가
func (r *UserRepository) AddLoginHistory(history *models.LoginHistory) error {
	history.CreatedAt = time.Now().Unix()
	history.UpdatedAt = history.CreatedAt
	_, err := r.db.Collection("login_history").InsertOne(context.Background(), history)
	return err
}

// 가장 최근 로그인 이력 가져오기
func (r *UserRepository) GetLatestLoginHistory(userID primitive.ObjectID) (*models.LoginHistory, error) {
	var history models.LoginHistory
	err := r.db.Collection("login_history").FindOne(
		context.Background(),
		bson.M{"user_id": userID},
	).Decode(&history)
	return &history, err
}

// 로그인 이력 업데이트
func (r *UserRepository) UpdateLoginHistory(id primitive.ObjectID, accessToken string) error {
	_, err := r.db.Collection("login_history").UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"access_token": accessToken, "updated_at": time.Now().Unix()}},
	)
	return err
}

// 사용자 저장
func (r *UserRepository) CreateUser(user *models.User) error {
	user.CreatedAt = time.Now().Unix()
	user.UpdatedAt = user.CreatedAt
	result, err := r.db.Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return err
}

func (r *UserRepository) AddEmailVerification(verification *models.EmailVerification) error {
	verification.ExpiresAt = time.Now().Add(15 * time.Minute).Unix()
	_, err := r.db.Collection("email_verifications").InsertOne(context.Background(), verification)

	// 에러가 발생하면 로그 출력
	if err != nil {
		log.Printf("Failed to insert email verification: %v\n", err)
	}

	return err
}

// 이메일 인증 토큰으로 인증 이력 조회
func (r *UserRepository) FindEmailVerification(token string) (*models.EmailVerification, error) {
	var verification models.EmailVerification
	err := r.db.Collection("email_verifications").FindOne(
		context.Background(),
		bson.M{"token": token},
	).Decode(&verification)
	return &verification, err
}

func (r *UserRepository) MarkEmailVerified(userID primitive.ObjectID) error {
	result, err := r.db.Collection("users").UpdateOne(
		context.Background(),
		bson.M{"_id": userID},
		bson.M{
			"$set": bson.M{"is_email_verified": true,
				"updated_at": time.Now().Unix(),
			}},
	)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no document found with userID: %s", userID.Hex())
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("document found but was not modified: %s", userID.Hex())
	}

	return nil
}

func (r *UserRepository) GetUserByID(userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.db.Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
