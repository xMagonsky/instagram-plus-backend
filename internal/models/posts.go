package models

import "time"

type AddPostRequest struct {
	Author  string `json:"author" binding:"required"`
	Image   string `json:"image" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type Post struct { // to change
	ID              int       `json:"id"`
	PhotoURL        string    `json:"image"`
	Description     string    `json:"content"`
	CreateTimestamp time.Time `json:"create_timestamp"`
	CreatorID       string    `json:"creator_id"`
	Author          string    `json:"author"`
}
