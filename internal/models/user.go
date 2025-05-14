package models

import "time"

type Profile struct {
	Username          string    `json:"username"`
	Name              string    `json:"name"`
	Surname           string    `json:"surname"`
	Description       string    `json:"description"`
	ProfileImageURL   string    `json:"profile_image_url"`
	Gender            string    `json:"gender"`
	BirthDate         time.Time `json:"birth_date"`
	CreationTimestamp time.Time `json:"creation_timestamp"`
	FollowersCount    int       `json:"followers_count"`
	FollowingCount    int       `json:"following_count"`
}

type UpdateProfileRequest struct {
	Name        string `json:"name" binding:"omitempty,max=20"`
	Surname     string `json:"surname" binding:"omitempty,max=20"`
	Description string `json:"description" binding:"omitempty,max=255"`
	Gender      string `json:"gender" binding:"omitempty,oneof=male female other"`
}
