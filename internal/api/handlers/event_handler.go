package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"district-friends/internal/models"
)

type EventHandler struct {
	DB *gorm.DB
}

func NewEventHandler(db *gorm.DB) *EventHandler {
	return &EventHandler{DB: db}
}

// @Summary List Events
// @Description Get a list of mock events (movies, concerts, dining)
// @Tags events
// @Produce json
// @Param type query string false "Filter by event type (e.g. Movie, Concert)"
// @Success 200 {array} models.Event
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/events [get]
func (h *EventHandler) ListEvents(c *gin.Context) {
	var events []models.Event
	query := h.DB

	eventType := c.Query("type")
	if eventType != "" {
		query = query.Where("type = ?", eventType)
	}

	if err := query.Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// @Summary Get Event Details
// @Description Get details for a specific event
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} models.Event
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := h.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}
