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
	Username       string `json:"username" binding:"required"`
	FriendUsername string `json:"friend_username" binding:"required"`
}

// @Summary Send a friend request
// @Description Send a friend request to another user using mobile numbers
// @Tags friends
// @Accept json
// @Produce json
// @Param request body FriendRequestReq true "Friend Request Info"
// @Success 201 {object} models.Friendship
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/request [post]
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	var req FriendRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.FriendUsername = req.FriendUsername

	if req.Username == req.FriendUsername {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send friend request to yourself"})
		return
	}

	var user, friend models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}
	if err := h.DB.Where("username = ?", req.FriendUsername).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ?", user.ID, friend.ID).First(&friendship).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend request already sent or users are already friends"})
		return
	}

	// Check if the other user already sent a request to this user
	var reciprocal models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", friend.ID, user.ID, models.StatusPending).First(&reciprocal).Error; err == nil {
		// Auto-accept the reciprocal request
		reciprocal.Status = models.StatusAccepted
		h.DB.Save(&reciprocal)

		// Create the new friendship as accepted
		friendship = models.Friendship{
			UserID:   user.ID,
			FriendID: friend.ID,
			Status:   models.StatusAccepted,
		}
		h.DB.Create(&friendship)

		c.JSON(http.StatusCreated, friendship)
		return
	}

	friendship = models.Friendship{
		UserID:   user.ID,
		FriendID: friend.ID,
		Status:   models.StatusPending,
	}

	if err := h.DB.Create(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send friend request"})
		return
	}

	c.JSON(http.StatusCreated, friendship)
}

// @Summary Accept a friend request
// @Description Accept a pending friend request using mobile numbers
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

	req.FriendUsername = req.FriendUsername

	var user, friend models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}
	if err := h.DB.Where("username = ?", req.FriendUsername).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	var friendship models.Friendship
	// Find the request that was sent BY friend TO user
	if err := h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", friend.ID, user.ID, models.StatusPending).First(&friendship).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found or not pending"})
		return
	}

	friendship.Status = models.StatusAccepted
	if err := h.DB.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept friend request"})
		return
	}

	// Create or update the reciprocal record so friendship is bidirectional
	var reciprocal models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ?", user.ID, friend.ID).First(&reciprocal).Error; err != nil {
		reciprocal = models.Friendship{
			UserID:   user.ID,
			FriendID: friend.ID,
			Status:   models.StatusAccepted,
		}
		h.DB.Create(&reciprocal)
	} else {
		reciprocal.Status = models.StatusAccepted
		h.DB.Save(&reciprocal)
	}

	c.JSON(http.StatusOK, friendship)
}

// @Summary Reject a friend request
// @Description Reject a pending friend request using mobile numbers
// @Tags friends
// @Accept json
// @Produce json
// @Param request body FriendRequestReq true "Friend Request Info"
// @Success 200 {object} models.Friendship
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/reject [post]
func (h *FriendHandler) RejectFriendRequest(c *gin.Context) {
	var req FriendRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.FriendUsername = req.FriendUsername

	var user, friend models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}
	if err := h.DB.Where("username = ?", req.FriendUsername).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	var friendship models.Friendship
	// Find the request that was sent BY friend TO user
	if err := h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", friend.ID, user.ID, models.StatusPending).First(&friendship).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found or not pending"})
		return
	}

	// Delete it instead of declining so they can send another one later if they want
	if err := h.DB.Delete(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject friend request"})
		return
	}

	// Delete any pending reciprocal request just in case they both sent one
	h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", user.ID, friend.ID, models.StatusPending).Delete(&models.Friendship{})

	c.JSON(http.StatusOK, gin.H{"message": "Friend request rejected and deleted"})
}

// @Summary List sent friend requests
// @Description Retrieve a list of pending friend requests sent by the user
// @Tags friends
// @Produce json
// @Param username query string true "User Mobile Number"
// @Success 200 {array} FriendResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/requests/sent [get]
func (h *FriendHandler) ListSentFriendRequests(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var friendships []models.Friendship
	if err := h.DB.Preload("Friend").Where("user_id = ? AND status = ?", user.ID, models.StatusPending).Find(&friendships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sent requests"})
		return
	}

	requests := make([]FriendResponse, 0)
	for _, f := range friendships {
		requests = append(requests, FriendResponse{
			User:   f.Friend,
			Status: string(f.Status),
		})
	}

	c.JSON(http.StatusOK, requests)
}

// @Summary List received friend requests
// @Description Retrieve a list of pending friend requests received by the user
// @Tags friends
// @Produce json
// @Param username query string true "User Mobile Number"
// @Success 200 {array} FriendResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/requests/received [get]
func (h *FriendHandler) ListReceivedFriendRequests(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var friendships []models.Friendship
	// Notice we check where friend_id matches our user ID, and preload the User who sent it.
	if err := h.DB.Preload("User").Where("friend_id = ? AND status = ?", user.ID, models.StatusPending).Find(&friendships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve received requests"})
		return
	}

	requests := make([]FriendResponse, 0)
	for _, f := range friendships {
		requests = append(requests, FriendResponse{
			User:   f.User,
			Status: string(f.Status),
		})
	}

	c.JSON(http.StatusOK, requests)
}

type FriendResponse struct {
	models.User
	Status string `json:"status"`
}

// @Summary List friends
// @Description Retrieve a list of accepted friends for a user by mobile number
// @Tags friends
// @Produce json
// @Param username query string true "User Mobile Number"
// @Success 200 {array} FriendResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends [get]
func (h *FriendHandler) ListFriends(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	var user models.User
	if err := h.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var friendships []models.Friendship
	if err := h.DB.Preload("Friend").Where("user_id = ? AND status = ?", user.ID, models.StatusAccepted).Find(&friendships).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve friends"})
		return
	}

	friends := make([]FriendResponse, 0)
	for _, f := range friendships {
		friends = append(friends, FriendResponse{
			User:   f.Friend,
			Status: string(f.Status),
		})
	}

	c.JSON(http.StatusOK, friends)
}

// @Summary Get friend status
// @Description Get the current friendship status between two users via mobile numbers
// @Tags friends
// @Produce json
// @Param username query string true "User Mobile Number"
// @Param friend_phone query string true "Friend Mobile Number"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/friends/status [get]
func (h *FriendHandler) GetFriendStatus(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}
	friendUsername := c.Query("friend_username")
	if friendUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "friend_username is required"})
		return
	}

	var user, friend models.User
	if err := h.DB.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err := h.DB.Where("username = ?", friendUsername).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend not found"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ?", user.ID, friend.ID).First(&friendship).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "none"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": string(friendship.Status)})
}
