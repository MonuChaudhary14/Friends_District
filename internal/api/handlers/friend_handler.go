package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"district-friends/internal/models"
)

type FriendHandler struct {
	DB *gorm.DB
}

func NewFriendHandler(db *gorm.DB) *FriendHandler {
	return &FriendHandler{DB: db}
}

type FriendRequestReq struct {
	UserID   uint `json:"user_id" binding:"required"`
	FriendID uint `json:"friend_id" binding:"required"`
}

// @Summary Send a friend request
// @Description Send a friend request to another user
// @Tags friends
// @Accept json
// @Produce json
// @Param request body FriendRequestReq true "Friend Request Info"
// @Success 201 {object} models.Friendship
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/request [post]
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	var req FriendRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == req.FriendID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send friend request to yourself"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ?", req.UserID, req.FriendID).First(&friendship).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend request already sent or users are already friends"})
		return
	}

	friendship = models.Friendship{
		UserID:   req.UserID,
		FriendID: req.FriendID,
		Status:   models.StatusPending,
	}

	if err := h.DB.Create(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send friend request"})
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// @Summary Accept a friend request
// @Description Accept a pending friend request
// @Tags friends
// @Accept json
// @Produce json
// @Param request body FriendRequestReq true "Friend Request Info"
// @Success 200 {object} models.Friendship
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/accept [post]
func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	var req FriendRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", req.UserID, req.FriendID, models.StatusPending).First(&friendship).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found or not pending"})
		return
	}

	friendship.Status = models.StatusAccepted
	if err := h.DB.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept friend request"})
		return
	}

	c.JSON(http.StatusOK, friendship)
}

// @Summary List friends
// @Description Retrieve a list of accepted friends for a user
// @Tags friends
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends [get]
func (h *FriendHandler) ListFriends(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	var friendships []models.Friendship
	if err := h.DB.Preload("Friend").Where("user_id = ? AND status = ?", userID, models.StatusAccepted).Find(&friendships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve friends"})
		return
	}

	friends := make([]models.User, 0)
	for _, f := range friendships {
		friends = append(friends, f.Friend)
	}

	c.JSON(http.StatusOK, friends)
}
