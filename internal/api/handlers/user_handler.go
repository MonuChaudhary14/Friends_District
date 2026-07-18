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

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type CreateProfileRequest struct {
	Name         string `json:"name" binding:"required"`
	Email        string `json:"email" binding:"required,email"`
	MobileNumber string `json:"mobile_number" binding:"required"`
	Username     string `json:"username,omitempty"` // Optional for profile creation
}

// @Summary Create a user
// @Description Register a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User Info"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Username: req.Username,
		Email:    req.Email,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
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
