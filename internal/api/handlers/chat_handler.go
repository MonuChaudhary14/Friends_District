package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"district-friends/internal/models"
)

type ChatHandler struct {
	DB       *gorm.DB
	Upgrader websocket.Upgrader
}

func NewChatHandler(db *gorm.DB) *ChatHandler {
	return &ChatHandler{
		DB: db,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

type CreateRoomReq struct {
	Name string `json:"name" binding:"required"`
}

// @Summary Create a chat room
// @Description Create a new chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param request body CreateRoomReq true "Room Name"
// @Success 201 {object} models.ChatRoom
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms [post]
func (h *ChatHandler) CreateRoom(c *gin.Context) {
	var req CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := models.ChatRoom{Name: req.Name}
	if err := h.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, room)
}

type JoinRoomReq struct {
	UserID uint `json:"user_id" binding:"required"`
}

// @Summary Join a chat room
// @Description Add a user to a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body JoinRoomReq true "User ID"
// @Success 201 {object} models.ChatRoomMember
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/{id}/join [post]
func (h *ChatHandler) JoinRoom(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var req JoinRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	member := models.ChatRoomMember{
		RoomID:   uint(roomID),
		UserID:   req.UserID,
		JoinedAt: time.Now(),
	}

	if err := h.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join room"})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// @Summary Get room messages
// @Description Retrieve message history for a chat room
// @Tags chat
// @Produce json
// @Param id path int true "Room ID"
// @Success 200 {array} models.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/{id}/messages [get]
func (h *ChatHandler) GetMessages(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var messages []models.Message
	if err := h.DB.Where("room_id = ?", roomID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// @Summary WebSocket for Chat Room
// @Description Connect to real-time chat via WebSocket
// @Tags chat
// @Param id path int true "Room ID"
// @Param user_id query int true "User ID"
// @Router /api/v1/rooms/{id}/ws [get]
func (h *ChatHandler) ServeWS(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	userIDStr := c.Query("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	conn, err := h.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		_, msgData, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		msg := models.Message{
			RoomID:    uint(roomID),
			SenderID:  uint(userID),
			Content:   string(msgData),
			CreatedAt: time.Now(),
		}

		if err := h.DB.Create(&msg).Error; err != nil {
			log.Println("Database error:", err)
			continue
		}

		// Simple echo to the current user. In a real app we'd broadcast to a Hub.
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Content)); err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}

type ShareEventReq struct {
	UserID            uint   `json:"user_id" binding:"required"`
	ExternalEventID   string `json:"external_event_id" binding:"required"`
	ExternalEventType string `json:"external_event_type" binding:"required"`
}

// @Summary Share an external event in a chat room
// @Description Share a specific TMDB or Ticketmaster event to a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body ShareEventReq true "User ID and External Event Details"
// @Success 201 {object} models.Message
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/{id}/share [post]
func (h *ChatHandler) ShareEvent(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var req ShareEventReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg := models.Message{
		RoomID:            uint(roomID),
		SenderID:          req.UserID,
		ExternalEventID:   req.ExternalEventID,
		ExternalEventType: req.ExternalEventType,
		Content:           "Check out this " + req.ExternalEventType + " event!",
		CreatedAt:         time.Now(),
	}

	if err := h.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to share event"})
		return
	}

	c.JSON(http.StatusCreated, msg)
}
