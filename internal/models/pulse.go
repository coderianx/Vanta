package models

import "time"

type CreateNewPulseRequest struct {
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// Pulse represents a published pulse and its database metadata.
type Pulse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}
