package models

type ExternalEvent struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Type        string  `json:"type"` // "movie" or "concert"
	Location    string  `json:"location,omitempty"`
	Date        string  `json:"date,omitempty"` // String format for simplicity
	ImageURL    string  `json:"image_url,omitempty"`
	PriceMin    float64 `json:"price_min,omitempty"`
	PriceMax    float64 `json:"price_max,omitempty"`
	URL         string  `json:"url,omitempty"`
}
