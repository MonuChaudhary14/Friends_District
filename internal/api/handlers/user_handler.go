package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"district-friends/internal/models"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

type CreateProfileRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	MobileNumber string `json:"mobile_number" binding:"required"`
	Username     string `json:"username,omitempty"` // Optional for profile creation
}

// @Summary Create a user profile
// @Description Register a new profile with name, email, and mobile number
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateProfileRequest true "Profile Info"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/profile [post]
func (h *UserHandler) CreateProfile(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username := req.Username
	if username == "" {
		firstName := strings.Split(strings.TrimSpace(req.Name), " ")[0]
		firstName = strings.ToLower(firstName)
		rand.Seed(time.Now().UnixNano())
		randomNumber := rand.Intn(9000) + 1000 // 4 digit random number
		username = fmt.Sprintf("%s%d", firstName, randomNumber)
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		MobileNumber: req.MobileNumber,
		Username:     username,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// @Summary Get a user profile
// @Description Retrieve a user's profile using their mobile number
// @Tags users
// @Produce json
// @Param mobile_number query string true "Mobile Number"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	mobileNumber := c.Query("mobile_number")
	if mobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mobile_number is required"})
		return
	}

	var user models.User
	if err := h.DB.Where("mobile_number = ?", mobileNumber).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}
