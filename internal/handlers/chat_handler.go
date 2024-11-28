package handlers

import (
	"chat-go-api/internal/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// CreateChatRoomHandler 채팅방 생성
func (h *ChatHandler) CreateChatRoomHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string   `json:"name"`
		MemberIDs []string `json:"member_ids"`
	}

	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 요청 데이터 파싱
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// MemberIDs를 ObjectID로 변환
	memberIDs := []primitive.ObjectID{}
	for _, id := range req.MemberIDs {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, "Invalid member ID", http.StatusBadRequest)
			return
		}
		memberIDs = append(memberIDs, objID)
	}

	// 채팅방 생성
	room, err := h.chatService.CreateChatRoom(req.Name, memberIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 응답
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

// GetUserChatRoomsHandler 유저가 참여한 채팅방 목록 조회
func (h *ChatHandler) GetUserChatRoomsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id")
	if userID == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	rooms, err := h.chatService.GetUserChatRooms(userID.(string))
	if err != nil {
		http.Error(w, "Failed to get chat rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

func (h *ChatHandler) GetChatRoomMessagesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomID"]

	limitParam := r.URL.Query().Get("limit")
	pageParam := r.URL.Query().Get("page")

	limit := int64(20)
	if limitParam != "" {
		parsedLimit, err := strconv.ParseInt(limitParam, 10, 64)
		if err != nil || parsedLimit <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = parsedLimit
	}

	page := int64(1)
	if pageParam != "" {
		parsedPage, err := strconv.ParseInt(pageParam, 10, 64)
		if err != nil || parsedPage <= 0 {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		page = parsedPage
	}

	messages, err := h.chatService.GetRecentMessages(roomID, limit, page)
	if err != nil {
		log.Printf("Failed to get messages: %v", err)
		http.Error(w, "failed to retrieve messages", http.StatusInternalServerError)
		return
	}

	// 메시지 총 개수 가져오기
	totalMessages, err := h.chatService.GetTotalMessagesCount(roomID)
	if err != nil {
		log.Printf("Failed to get total message count: %v", err)
		http.Error(w, "failed to retrieve total message count", http.StatusInternalServerError)
		return
	}

	// 응답 생성
	response := map[string]interface{}{
		"total":    totalMessages,
		"page":     page,
		"limit":    limit,
		"messages": messages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ChatHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("", h.GetUserChatRoomsHandler).Methods("GET")
	router.HandleFunc("", h.CreateChatRoomHandler).Methods("POST")
	router.HandleFunc("/{roomID}/messages", h.GetChatRoomMessagesHandler).Methods("GET")
}
