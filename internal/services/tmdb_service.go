package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"district-friends/internal/models"
)

type TMDBService struct {
	APIKey string
}

func NewTMDBService() *TMDBService {
	return &TMDBService{
		APIKey: os.Getenv("TMDB_API_Key"),
	}
}

type TMDBResponse struct {
	Results []struct {
		ID               int    `json:"id"`
		Title            string `json:"title"`
		Overview         string `json:"overview"`
		ReleaseDate      string `json:"release_date"`
		PosterPath       string `json:"poster_path"`
	} `json:"results"`
}

func (s *TMDBService) FetchMovies() ([]models.ExternalEvent, error) {
	if s.APIKey == "" {
		return nil, fmt.Errorf("TMDB_API_Key not configured")
	}

	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/popular?api_key=%s&language=en-US&page=1", s.APIKey)

	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch TMDB data: status %d", resp.StatusCode)
	}

	var tmdbResp TMDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&tmdbResp); err != nil {
		return nil, err
	}

	var events []models.ExternalEvent
	for _, movie := range tmdbResp.Results {
		events = append(events, models.ExternalEvent{
			ID:          fmt.Sprintf("%d", movie.ID),
			Title:       movie.Title,
			Description: movie.Overview,
			Type:        "movie",
			Date:        movie.ReleaseDate,
			ImageURL:    "https://image.tmdb.org/t/p/w500" + movie.PosterPath,
		})
	}

	return events, nil
}
