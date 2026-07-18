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
	UserID            uint    `json:"user_id" binding:"required"`
	BookedForID       *uint   `json:"booked_for_id"` // Optional
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

	// If booking for someone else, verify friendship
	if req.BookedForID != nil {
		var friendship models.Friendship
		err := h.DB.Where(
			"((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)) AND status = ?",
			req.UserID, *req.BookedForID, *req.BookedForID, req.UserID, models.StatusAccepted,
		).First(&friendship).Error

		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only book tickets for confirmed friends"})
			return
		}
	}

	booking := models.Booking{
		UserID:            req.UserID,
		BookedForID:       req.BookedForID,
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
// @Param user_id query int true "User ID"
// @Success 200 {array} models.Booking
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/bookings [get]
func (h *BookingHandler) ListBookings(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id"})
		return
	}

	var bookings []models.Booking
	if err := h.DB.Where("user_id = ? OR booked_for_id = ?", userID, userID).
		Preload("User").Preload("BookedFor").Find(&bookings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
