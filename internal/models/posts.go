package models

import "time"

type Post struct {
	ID                    int       `json:"id"`
	AuthorUsername        string    `json:"author_username"`
	ImageURL              string    `json:"image_url"`
	Description           string    `json:"description"`
	CreationTimestamp     time.Time `json:"create_timestamp"`
	AuthorName            string    `json:"author_name"`
	AuthorSurname         string    `json:"author_surname"`
	LikesCount            int       `json:"likes_count"`
	AlreadyLiked          bool      `json:"already_liked"`
	AuthorProfileImageURL string    `json:"author_profile_image_url"`
	CommentsCount         int       `json:"comments_count"`
}

type AddPostRequest struct {
	Description string `json:"description" binding:"required,max=255"`
}

type UpdatePostRequest struct {
	Description string `json:"description" binding:"required,max=255"`
}
