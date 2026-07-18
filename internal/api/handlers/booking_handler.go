package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"district-friends/internal/models"
)

type BookingHandler struct {
	DB *gorm.DB
}

func NewBookingHandler(db *gorm.DB) *BookingHandler {
	return &BookingHandler{DB: db}
}

type CreateBookingReq struct {
	UserPhone         string  `json:"user_phone" binding:"required"`
	BookedForPhone    *string `json:"booked_for_phone"` // Optional
	ExternalEventID   string  `json:"external_event_id" binding:"required"`
	ExternalEventType string  `json:"external_event_type" binding:"required"`
	Quantity          int     `json:"quantity" binding:"required"`
	TotalPrice        float64 `json:"total_price"`
}

// @Summary Create a Booking
// @Description Book a ticket for yourself or a friend
// @Tags bookings
// @Accept json
// @Produce json
// @Param request body CreateBookingReq true "Booking Details"
// @Success 201 {object} models.Booking
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req CreateBookingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", req.UserPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var bookedForID *uint
	// If booking for someone else, verify friendship
	if req.BookedForPhone != nil {
		var friend models.User
		if err := h.DB.Where("mobile_number = ?", *req.BookedForPhone).First(&friend).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Booked for user not found"})
			return
		}
		bookedForID = &friend.ID
		var friendship models.Friendship
		err := h.DB.Where(
			"((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)) AND status = ?",
			user.ID, *bookedForID, *bookedForID, user.ID, models.StatusAccepted,
		).First(&friendship).Error

		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only book tickets for confirmed friends"})
			return
		}
	}

	booking := models.Booking{
		UserID:            user.ID,
		BookedForID:       bookedForID,
		ExternalEventID:   req.ExternalEventID,
		ExternalEventType: req.ExternalEventType,
		Quantity:          req.Quantity,
		TotalPrice:        req.TotalPrice,
		Status:            "CONFIRMED",
		CreatedAt:         time.Now(),
	}

	if err := h.DB.Create(&booking).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

// @Summary List Bookings
// @Description Get a list of bookings for a specific user (either bought by them, or bought for them)
// @Tags bookings
// @Produce json
// @Param user_phone query string true "User Phone"
// @Success 200 {array} models.Booking
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/bookings [get]
func (h *BookingHandler) ListBookings(c *gin.Context) {
	userPhone := c.Query("user_phone")
	if userPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_phone is required"})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", userPhone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var bookings []models.Booking
	if err := h.DB.Where("user_id = ? OR booked_for_id = ?", user.ID, user.ID).
		Preload("User").Preload("BookedFor").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
