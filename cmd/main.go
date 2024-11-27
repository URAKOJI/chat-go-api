package main

import (
	"chat-go-api/internal/handlers"
	"chat-go-api/internal/repository"
	"chat-go-api/internal/services"
	"chat-go-api/internal/websocket"
	"chat-go-api/pkg/utils"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// .env 파일 로드
	err := godotenv.Load()
	if err != nil {
		log.Println(".env file not found, using default values from config.yaml")
	}

	// config.yaml 파일 로드
	config, err := utils.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// .env 파일 값으로 config 덮어쓰기
	overrideConfigWithEnv(config)

	// DB 연결
	db, err := repository.ConnectDB(config.Database.Url, "chat_db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// WebSocket 매니저 초기화
	wsManager := websocket.NewManager()
	go wsManager.Run() // WebSocket 매니저 실행

	// EmailService 초기화
	emailService := services.NewEmailService(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)

	// AuthService 초기화
	userRepo := repository.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, emailService, "access-secret-key", "refresh-secret-key")

	// 핸들러 초기화
	authHandler := handlers.NewAuthHandler(authService)

	// 라우터 설정
	router := mux.NewRouter()
	authHandler.RegisterRoutes(router) // 회원가입 및 인증 관련 라우트 추가
	router.HandleFunc("/ws", websocket.WebSocketHandler(wsManager))

	// 서버 시작
	fmt.Printf("Server starting on port %s\n", config.Server.Port)
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, router))
}

// .env 값을 config에 적용하는 함수
func overrideConfigWithEnv(config *utils.Config) {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		config.Database.Url = dbURL
	}
}
