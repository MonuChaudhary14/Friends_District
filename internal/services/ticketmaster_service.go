package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"district-friends/internal/models"
)

type TicketmasterService struct {
	APIKey string
}

func NewTicketmasterService() *TicketmasterService {
	return &TicketmasterService{
		APIKey: os.Getenv("Ticketmaster_Consumer_Key"),
	}
}

type TicketmasterResponse struct {
	Embedded struct {
		Events []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			URL    string `json:"url"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			Dates struct {
				Start struct {
					LocalDate string `json:"localDate"`
					LocalTime string `json:"localTime"`
				} `json:"start"`
			} `json:"dates"`
			Embedded struct {
				Venues []struct {
					Name string `json:"name"`
				} `json:"venues"`
			} `json:"_embedded"`
			PriceRanges []struct {
				Min float64 `json:"min"`
				Max float64 `json:"max"`
			} `json:"priceRanges"`
		} `json:"events"`
	} `json:"_embedded"`
}

func (s *TicketmasterService) FetchConcerts() ([]models.ExternalEvent, error) {
	if s.APIKey == "" {
		return nil, fmt.Errorf("Ticketmaster_Consumer_Key not configured")
	}

	// Fetch upcoming music events (classificationName=music)
	url := fmt.Sprintf("https://app.ticketmaster.com/discovery/v2/events.json?apikey=%s&classificationName=music&size=20", s.APIKey)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch Ticketmaster data: status %d", resp.StatusCode)
	}

	var tmResp TicketmasterResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmResp); err != nil {
		return nil, err
	}

	var events []models.ExternalEvent
	for _, event := range tmResp.Embedded.Events {
		imageURL := ""
		if len(event.Images) > 0 {
			imageURL = event.Images[0].URL
		}

		location := ""
		if len(event.Embedded.Venues) > 0 {
			location = event.Embedded.Venues[0].Name
		}

		priceMin, priceMax := 0.0, 0.0
		if len(event.PriceRanges) > 0 {
			priceMin = event.PriceRanges[0].Min
			priceMax = event.PriceRanges[0].Max
		}

		events = append(events, models.ExternalEvent{
			ID:          event.ID,
			Title:       event.Name,
			Description: "Live concert at " + location,
			Type:        "concert",
			Location:    location,
			Date:        event.Dates.Start.LocalDate + " " + event.Dates.Start.LocalTime,
			ImageURL:    imageURL,
			PriceMin:    priceMin,
			PriceMax:    priceMax,
			URL:         event.URL,
		})
	}

	return events, nil
}
