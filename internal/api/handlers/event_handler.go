package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"district-friends/internal/models"
	"district-friends/internal/services"
)

type EventHandler struct {
	TMDB         *services.TMDBService
	Ticketmaster *services.TicketmasterService
	Foursquare   *services.FoursquareService
}

func NewEventHandler() *EventHandler {
	return &EventHandler{
		TMDB:         services.NewTMDBService(),
		Ticketmaster: services.NewTicketmasterService(),
		Foursquare:   services.NewFoursquareService(),
	}
}

// @Summary List Events
// @Description Get a list of mock events (movies, concerts, dining)
// @Tags events
// @Produce json
// @Param type query string false "Filter by event type (e.g. Movie, Concert)"
// @Success 200 {array} models.ExternalEvent
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/events [get]
func (h *EventHandler) ListEvents(c *gin.Context) {
	var allEvents = []models.ExternalEvent{}

	eventType := c.Query("type")

	// Helper function to deduplicate
	seen := make(map[string]bool)
	addEvents := func(events []models.ExternalEvent) {
		for _, e := range events {
			if !seen[e.Title] {
				seen[e.Title] = true
				allEvents = append(allEvents, e)
			}
		}
	}

	// Fetch Movies
	if eventType == "" || eventType == "movie" {
		movies, err := h.TMDB.FetchMovies()
		if err == nil {
			addEvents(movies)
		} else {
			log.Printf("TMDB Error: %v", err)
		}
	}

	// Fetch Concerts
	if eventType == "" || eventType == "concert" {
		concerts, err := h.Ticketmaster.FetchConcerts()
		if err == nil {
			addEvents(concerts)
		} else {
			log.Printf("Ticketmaster Error: %v", err)
		}
	}

	// Fetch Dining
	if eventType == "" || eventType == "dining" || eventType == "restaurant" {
		dining, err := h.Foursquare.FetchDining()
		if err == nil {
			addEvents(dining)
		} else {
			log.Printf("Foursquare Error: %v", err)
		}
	}

	c.JSON(http.StatusOK, allEvents)
}

// @Summary Get Event Details
// @Description Get details for a specific event
// @Tags events
// @Produce json
// @Param id path int true "Event ID"
// @Success 200 {object} models.ExternalEvent
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/events/{id} [get]
func (h *EventHandler) GetEvent(c *gin.Context) {
	// For this prototype, since we merge data, fetching a single event by ID across 2 different APIs 
	// would require knowing which API the ID belongs to. The frontend usually passes the type, or we search both.
	// We'll leave a simple stub here that tells the client to use the list endpoint, or we can implement a search.
	
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Single event lookup is not fully implemented for merged external APIs. Use the list endpoint."})
}

// @Summary Get Spotlight Events
// @Description Get a curated list of top trending events (one movie, one concert, one dining)
// @Tags events
// @Produce json
// @Success 200 {array} models.ExternalEvent
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/events/spotlight [get]
func (h *EventHandler) GetSpotlight(c *gin.Context) {
	var spotlights []models.ExternalEvent

	movies, err := h.TMDB.FetchMovies()
	if err == nil && len(movies) > 0 {
		spotlights = append(spotlights, movies[0])
	} else {
		log.Printf("TMDB Error: %v", err)
	}

	concerts, err := h.Ticketmaster.FetchConcerts()
	if err == nil && len(concerts) > 0 {
		spotlights = append(spotlights, concerts[0])
	} else {
		log.Printf("Ticketmaster Error: %v", err)
	}

	dining, err := h.Foursquare.FetchDining()
	if err == nil && len(dining) > 0 {
		spotlights = append(spotlights, dining[0])
	} else {
		log.Printf("Foursquare Error: %v", err)
	}

	c.JSON(http.StatusOK, spotlights)
}
