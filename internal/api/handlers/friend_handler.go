package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"district-friends/internal/models"
	"district-friends/internal/utils"
)

type FriendHandler struct {
	DB *gorm.DB
}

func NewFriendHandler(db *gorm.DB) *FriendHandler {
	return &FriendHandler{DB: db}
}

type FriendRequestReq struct {
	UserPhone   string `json:"user_phone" binding:"required"`
	FriendPhone string `json:"friend_phone" binding:"required"`
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

	normUserPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}
	req.UserPhone = normUserPhone

	normFriendPhone, err := utils.ValidateAndNormalizePhone(req.FriendPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend_phone: " + err.Error()})
		return
	}
	req.FriendPhone = normFriendPhone

	if req.UserPhone == req.FriendPhone {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot send friend request to yourself"})
		return
	}

	var user, friend models.User
	if err := h.DB.Where("mobile_number = ?", req.UserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}
	if err := h.DB.Where("mobile_number = ?", req.FriendPhone).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ?", user.ID, friend.ID).First(&friendship).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend request already sent or users are already friends"})
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

	normUserPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}
	req.UserPhone = normUserPhone

	normFriendPhone, err := utils.ValidateAndNormalizePhone(req.FriendPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend_phone: " + err.Error()})
		return
	}
	req.FriendPhone = normFriendPhone

	var user, friend models.User
	if err := h.DB.Where("mobile_number = ?", req.UserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}
	if err := h.DB.Where("mobile_number = ?", req.FriendPhone).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", user.ID, friend.ID, models.StatusPending).First(&friendship).Error; err != nil {
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

	normUserPhone, err := utils.ValidateAndNormalizePhone(req.UserPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}
	req.UserPhone = normUserPhone

	normFriendPhone, err := utils.ValidateAndNormalizePhone(req.FriendPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend_phone: " + err.Error()})
		return
	}
	req.FriendPhone = normFriendPhone

	var user, friend models.User
	if err := h.DB.Where("mobile_number = ?", req.UserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}
	if err := h.DB.Where("mobile_number = ?", req.FriendPhone).First(&friend).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	var friendship models.Friendship
	if err := h.DB.Where("user_id = ? AND friend_id = ? AND status = ?", user.ID, friend.ID, models.StatusPending).First(&friendship).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found or not pending"})
		return
	}

	friendship.Status = models.StatusDeclined
	if err := h.DB.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject friend request"})
		return
	}

	c.JSON(http.StatusOK, friendship)
}

// @Summary List sent friend requests
// @Description Retrieve a list of pending friend requests sent by the user
// @Tags friends
// @Produce json
// @Param user_phone query string true "User Mobile Number"
// @Success 200 {array} FriendResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/requests/sent [get]
func (h *FriendHandler) ListSentFriendRequests(c *gin.Context) {
	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	normUserPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normUserPhone).First(&user).Error; err != nil {
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
// @Param user_phone query string true "User Mobile Number"
// @Success 200 {array} FriendResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends/requests/received [get]
func (h *FriendHandler) ListReceivedFriendRequests(c *gin.Context) {
	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	normUserPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normUserPhone).First(&user).Error; err != nil {
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
// @Param user_phone query string true "User Mobile Number"
// @Success 200 {array} FriendResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/friends [get]
func (h *FriendHandler) ListFriends(c *gin.Context) {
	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	normUserPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", normUserPhone).First(&user).Error; err != nil {
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
// @Param user_phone query string true "User Mobile Number"
// @Param friend_phone query string true "Friend Mobile Number"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/friends/status [get]
func (h *FriendHandler) GetFriendStatus(c *gin.Context) {
	userPhone := c.Query("user_phone")
	friendPhone := c.Query("friend_phone")

	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}
	if friendPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "friend_phone is required"})
		return
	}

	normUserPhone, err := utils.ValidateAndNormalizePhone(userPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_phone: " + err.Error()})
		return
	}

	normFriendPhone, err := utils.ValidateAndNormalizePhone(friendPhone)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid friend_phone: " + err.Error()})
		return
	}

	var user, friend models.User
	if err := h.DB.Where("mobile_number = ?", normUserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if err := h.DB.Where("mobile_number = ?", normFriendPhone).First(&friend).Error; err != nil {
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
