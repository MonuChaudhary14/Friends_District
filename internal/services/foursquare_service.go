package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"district-friends/internal/models"
)

type FoursquareService struct {
	BaseURL string
	APIKey  string
}

func NewFoursquareService() *FoursquareService {
	return &FoursquareService{
		BaseURL: "https://places-api.foursquare.com",
		APIKey:  os.Getenv("Foursquare_SERVICE_API"),
	}
}

// FoursquareResponse matches the new FSQ OS Places API structure
type FoursquareResponse struct {
	Results []struct {
		FsqPlaceID string `json:"fsq_place_id"`
		Name       string `json:"name"`
		Location   struct {
			FormattedAddress string `json:"formatted_address"`
		} `json:"location"`
		Description string `json:"description"`
	} `json:"results"`
}

func (s *FoursquareService) FetchDining() ([]models.ExternalEvent, error) {
	if s.APIKey == "" {
		return nil, fmt.Errorf("Foursquare_SERVICE_API is not set")
	}

	url := fmt.Sprintf("%s/places/search?query=restaurant&near=New York&limit=10", s.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Add necessary authentication headers for the new Places API
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("X-Places-Api-Version", "2025-06-17")
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Foursquare API returned status: %d", resp.StatusCode)
	}

	var fsqResp FoursquareResponse
	if err := json.NewDecoder(resp.Body).Decode(&fsqResp); err != nil {
		return nil, err
	}

	var events []models.ExternalEvent
	for _, venue := range fsqResp.Results {
		desc := venue.Description
		if desc == "" {
			desc = "Experience fine dining at " + venue.Name
		}
		
		events = append(events, models.ExternalEvent{
			ID:          "fsq_" + venue.FsqPlaceID,
			Title:       venue.Name,
			Description: desc,
			Type:        "Dining",
			Date:        time.Now().AddDate(0, 0, 2).Format("2006-01-02"), // Mocking a future date for dining events
			Location:    venue.Location.FormattedAddress,
			ImageURL:    "https://images.unsplash.com/photo-1517248135467-4c7edcad34c4?auto=format&fit=crop&q=80&w=800", // Generic restaurant image
		})
	}

	return events, nil
}
