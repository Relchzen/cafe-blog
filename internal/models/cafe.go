package models

import "time"

type Cafe struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Rating    *int      `json:"rating"`
	Notes     string    `json:"notes"`
	PhotoURL  *string   `json:"photo_url,omitempty"`
	PhotoURLs []string  `json:"photo_urls,omitempty"`
	VisitDate *string   `json:"visit_date"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CafePhoto struct {
	ID           int       `json:"id"`
	CafeID       int       `json:"cafe_id"`
	PhotoURL     string    `json:"photo_url"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateCafeRequest struct {
	Name      string  `json:"name"`
	Location  string  `json:"location"`
	Rating    *int    `json:"rating"`
	Notes     string  `json:"notes"`
	VisitDate *string `json:"visit_date"`
	PhotoURL  *string `json:"photo_url,omitempty"`
}

type UpdateCafeRequest struct {
	Name      string  `json:"name"`
	Location  string  `json:"location"`
	Rating    *int    `json:"rating"`
	Notes     string  `json:"notes"`
	VisitDate *string `json:"visit_date"`
	PhotoURL  *string `json:"photo_url,omitempty"`
}
