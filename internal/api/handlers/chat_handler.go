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
	"district-friends/internal/utils"
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
	UserPhone string `json:"user_phone" binding:"required"`
}

// @Summary Join a chat room
// @Description Add a user to a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body JoinRoomReq true "User Phone"
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

	normPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}
	req.UserPhone = normPhone

	var user models.User
	if err := h.DB.Where("mobile_number = ?", req.UserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	member := models.ChatRoomMember{
		RoomID:   uint(roomID),
		UserID:   user.ID,
		JoinedAt: time.Now(),
	}

	if err := h.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join room"})
		return
	}

	c.JSON(http.StatusCreated, member)
}

type InviteRoomReq struct {
	InviterPhone string `json:"inviter_phone" binding:"required"`
	InviteePhone string `json:"invitee_phone" binding:"required"`
}

// @Summary Invite a user to a chat room
// @Description Invite a user to join an existing chat room. The inviter must already be a member.
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body InviteRoomReq true "Inviter and Invitee Phones"
// @Success 201 {object} models.ChatRoomMember
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/{id}/invite [post]
func (h *ChatHandler) InviteToRoom(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var req InviteRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Normalize phones
	normInviter, err := utils.ValidateAndNormalizePhone(req.InviterPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inviter_phone: " + err.Error()})
		return
	}
	
	normInvitee, err := utils.ValidateAndNormalizePhone(req.InviteePhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid invitee_phone: " + err.Error()})
		return
	}

	// Find both users
	var inviter, invitee models.User
	if err := h.DB.Where("mobile_number = ?", normInviter).First(&inviter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inviter not found"})
		return
	}
	if err := h.DB.Where("mobile_number = ?", normInvitee).First(&invitee).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invitee not found"})
		return
	}

	// Check if inviter is in the room
	var inviterMembership models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND user_id = ?", roomID, inviter.ID).First(&inviterMembership).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "You must be a member of the room to invite others"})
		return
	}

	// Check if invitee is already in the room
	var inviteeMembership models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND user_id = ?", roomID, invitee.ID).First(&inviteeMembership).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of this room"})
		return
	}

	member := models.ChatRoomMember{
		RoomID:   uint(roomID),
		UserID:   invitee.ID,
		Status:   "pending",
		JoinedAt: time.Now(),
	}

	if err := h.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to room"})
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
// @Param user_phone query string true "User Phone"
// @Router /api/v1/rooms/{id}/ws [get]
func (h *ChatHandler) ServeWS(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	normPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
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
			SenderID:  user.ID,
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
	UserPhone         string `json:"user_phone" binding:"required"`
	ExternalEventID   string `json:"external_event_id" binding:"required"`
	ExternalEventType string `json:"external_event_type" binding:"required"`
}

// @Summary Share an external event in a chat room
// @Description Share a specific TMDB or Ticketmaster event to a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body ShareEventReq true "User Phone and External Event Details"
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

	normPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}
	req.UserPhone = normPhone

	var user models.User
	if err := h.DB.Where("mobile_number = ?", req.UserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	msg := models.Message{
		RoomID:            uint(roomID),
		SenderID:          user.ID,
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

// @Summary List Joined Rooms
// @Description Get a list of chat rooms that the user is a member of
// @Tags chat
// @Produce json
// @Param user_phone query string true "User Phone"
// @Success 200 {array} models.ChatRoom
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms [get]
func (h *ChatHandler) ListJoinedRooms(c *gin.Context) {
	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	normPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var members []models.ChatRoomMember
	if err := h.DB.Where("user_id = ? AND status = ?", user.ID, "joined").Preload("Room").Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch joined rooms"})
		return
	}

	rooms := []models.ChatRoom{}
	for _, member := range members {
		rooms = append(rooms, member.Room)
	}

	c.JSON(http.StatusOK, rooms)
}

type HandleRoomInviteReq struct {
	UserPhone string `json:"user_phone" binding:"required"`
}

// @Summary Accept a room invite
// @Description Accept a pending invitation to join a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body HandleRoomInviteReq true "User Phone"
// @Success 200 {object} models.ChatRoomMember
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/{id}/accept [post]
func (h *ChatHandler) AcceptRoomInvite(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var req HandleRoomInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone"})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var membership models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND user_id = ? AND status = ?", roomID, user.ID, "pending").First(&membership).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending invite not found"})
		return
	}

	membership.Status = "joined"
	membership.JoinedAt = time.Now()
	
	if err := h.DB.Save(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept invite"})
		return
	}

	c.JSON(http.StatusOK, membership)
}

// @Summary Reject a room invite
// @Description Reject a pending invitation to join a chat room
// @Tags chat
// @Accept json
// @Produce json
// @Param id path int true "Room ID"
// @Param request body HandleRoomInviteReq true "User Phone"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/{id}/reject [post]
func (h *ChatHandler) RejectRoomInvite(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.Atoi(roomIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room ID"})
		return
	}

	var req HandleRoomInviteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone"})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var membership models.ChatRoomMember
	if err := h.DB.Where("room_id = ? AND user_id = ? AND status = ?", roomID, user.ID, "pending").First(&membership).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending invite not found"})
		return
	}
	
	if err := h.DB.Delete(&membership).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject invite"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Invite rejected"})
}

// @Summary List Pending Room Invites
// @Description Get a list of chat rooms that the user has been invited to but hasn't joined
// @Tags chat
// @Produce json
// @Param user_phone query string true "User Phone"
// @Success 200 {array} models.ChatRoom
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/rooms/invites [get]
func (h *ChatHandler) ListRoomInvites(c *gin.Context) {
	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	normPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone"})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var members []models.ChatRoomMember
	if err := h.DB.Where("user_id = ? AND status = ?", user.ID, "pending").Preload("Room").Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch room invites"})
		return
	}

	rooms := []models.ChatRoom{}
	for _, member := range members {
		rooms = append(rooms, member.Room)
	}

	c.JSON(http.StatusOK, rooms)
}
