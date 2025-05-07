package models

import "time"

type Post struct {
	ID                int       `json:"id"`
	AuthorID          string    `json:"author_id"`
	ImageURL          string    `json:"image_url"`
	Description       string    `json:"description"`
	CreationTimestamp time.Time `json:"create_timestamp"`
	AuthorName        string    `json:"author_name"`
}

type AddPostRequest struct {
	ImageURL    string `json:"image_url" binding:"required"`
	Description string `json:"description" binding:"required"`
}

type UpdatePostRequest struct {
	Description string `json:"description" binding:"required"`
}
