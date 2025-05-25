package models

import "time"

type Comment struct {
	ID                    int       `json:"id"`
	PostID                int       `json:"post_id"`
	AuthorID              int       `json:"author_id"`
	AuthorUsername        string    `json:"author_username"`
	AuthorProfileImageURL string    `json:"author_profile_image_url"`
	Content               string    `json:"content"`
	CreationTimestamp     time.Time `json:"creation_timestamp"`
}

type AddCommentRequest struct {
	Content string `json:"content" binding:"required,max=255"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,max=255"`
}
