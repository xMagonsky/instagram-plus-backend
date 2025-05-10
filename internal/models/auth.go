package models

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`

	Description  string    `json:"description" binding:"max=255"`
	ProfileImage string    `json:"profile_image" binding:"max=255"`
	Name         string    `json:"name" binding:"required,max=20"`
	Surname      string    `json:"surname" binding:"required,max=20"`
	Gender       string    `json:"gender" binding:"required,oneof=male female other"`
	BirthDate    time.Time `json:"birth_date" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenRequest struct {
	Token string `json:"token" binding:"required"`
}
