package services

import (
	"chat-go-api/internal/models"
	"chat-go-api/internal/repository"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo          *repository.UserRepository
	accessSecret  string
	refreshSecret string
	emailService  *EmailService // 이메일 서비스 추가
}

func NewAuthService(repo *repository.UserRepository, emailService *EmailService, accessSecret, refreshSecret string) *AuthService {
	return &AuthService{
		repo:          repo,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		emailService:  emailService, // 이메일 서비스 초기화
	}
}

// 로그인 처리
func (s *AuthService) Login(email, password string) (string, string, error) {
	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// 이메일 인증 여부 확인
	if !user.IsEmailVerified {
		if !user.IsEmailVerified {
			// 인증 이메일 다시 발송
			token, err := s.generateVerificationToken()
			if err != nil {
				return "", "", fmt.Errorf("failed to generate verification token: %v", err)
			}

			verification := &models.EmailVerification{
				UserID:    user.ID,
				Token:     token,
				CreatedAt: time.Now().Unix(),
				ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
			}

			if err := s.repo.AddEmailVerification(verification); err != nil {
				return "", "", fmt.Errorf("failed to save verification token: %v", err)
			}

			if err := s.emailService.SendVerificationEmailAsync(email, token); err != nil {
				return "", "", err
			}

			return "", "", errors.New("email not verified. verification email resent")
		}
	}

	// 비밀번호 검증
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid password")
	}

	// 토큰 생성
	accessToken, err := s.generateToken(user.ID.Hex(), s.accessSecret, 5*time.Hour)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := s.generateToken(user.ID.Hex(), s.refreshSecret, 8*time.Hour)
	if err != nil {
		return "", "", err
	}

	// 로그인 이력 저장
	history := &models.LoginHistory{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	if err := s.repo.AddLoginHistory(history); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// 토큰 재발급
func (s *AuthService) RefreshToken(userID string, oldRefreshToken string) (string, error) {
	id, _ := primitive.ObjectIDFromHex(userID)
	history, err := s.repo.GetLatestLoginHistory(id)
	if err != nil {
		return "", errors.New("no login history found")
	}

	// Refresh 토큰 검증
	if history.RefreshToken != oldRefreshToken {
		return "", errors.New("invalid refresh token")
	}

	// 새로운 Access 토큰 생성
	newAccessToken, err := s.generateToken(userID, s.accessSecret, 5*time.Minute)
	if err != nil {
		return "", err
	}

	// 이력 업데이트
	if err := s.repo.UpdateLoginHistory(history.ID, newAccessToken); err != nil {
		return "", err
	}

	return newAccessToken, nil
}

// 토큰 생성
func (s *AuthService) generateToken(userID string, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// 회원가입 로직
func (s *AuthService) Register(email, password, name string) error {
	// 이메일 중복 확인
	_, err := s.repo.FindByEmail(email)
	if err == nil {
		return errors.New("email already exists")
	}

	// 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// 사용자 생성
	user := &models.User{
		Email:           email,
		Password:        string(hashedPassword),
		Name:            name,
		Role:            "user",
		IsEmailVerified: false,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return err
	}

	// 이메일 인증 토큰 생성
	token, err := s.generateVerificationToken()
	if err != nil {
		return err
	}

	verification := &models.EmailVerification{
		UserID:    user.ID,
		Token:     token,
		CreatedAt: time.Now().Unix(),
	}
	if err := s.repo.AddEmailVerification(verification); err != nil {
		return err
	}

	// 이메일 전송
	if err := s.emailService.SendVerificationEmailAsync(email, token); err != nil {
		return err
	}
	return nil
}

// 이메일 인증 처리
func (s *AuthService) VerifyEmail(token string) error {
	verification, err := s.repo.FindEmailVerification(token)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// verification 구조체 출력
	fmt.Printf("Verification Data: %+v\n", verification)

	// 토큰 만료 확인
	if verification.ExpiresAt < time.Now().Unix() {
		return errors.New("token expired")
	}

	// 사용자 이메일 인증 완료
	if err := s.repo.MarkEmailVerified(verification.UserID); err != nil {
		return err
	}

	return nil
}

// 이메일 인증 토큰 생성
func (s *AuthService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
