package routes

import (
	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (r *RoutesManager) RegisterUserRoutes(router *gin.Engine) {
	userRouter := router.Group("/profile")
	userRouter.Use(r.middleware.RequireAuth())
	{
		userRouter.GET("/:id", func(c *gin.Context) {
			userID := c.Param("id")
			if userID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
				return
			}

			var user models.Profile
			err := r.pgClient.QueryRow(c.Request.Context(), `
				SELECT u.username, p.name, p.surname, p.description, p.profile_image_url, p.gender, p.birth, u.creation_timestamp
				FROM users u
				JOIN user_profiles p ON u.id = p.user_id
				WHERE u.id = $1`, userID).Scan(
				&user.Username, &user.Name, &user.Surname, &user.Description, &user.ProfileImageURL, &user.Gender, &user.BirthDate, &user.CreationTimestamp)
			if err != nil {
				if err == pgx.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				} else {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				}
				return
			}

			c.JSON(http.StatusOK, user)
		})

		userRouter.PUT("/:id", r.middleware.RequireUserOwnership("id"), func(c *gin.Context) {
			userID := c.Param("id")
			if userID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
				return
			}

			var req models.UpdateProfileRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			if req.Name == "" && req.Surname == "" && req.Description == "" && req.Gender == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "at least one field is required"})
				return
			}

			_, err := r.pgClient.Exec(c.Request.Context(), `
				UPDATE user_profiles 
				SET name = COALESCE(NULLIF($1, ''), name), 
					surname = COALESCE(NULLIF($2, ''), surname),
					description = COALESCE(NULLIF($3, ''), description),
					gender = COALESCE(NULLIF($4, '')::gender, gender)
				WHERE user_id = $5`,
				req.Name, req.Surname, req.Description, req.Gender, userID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{})
		})
	}
}
