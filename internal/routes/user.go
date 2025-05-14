package routes

import (
	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (r *RoutesManager) RegisterUserRoutes(router *gin.Engine) {
	profileRouter := router.Group("/profile")
	profileRouter.Use(r.middleware.RequireAuth())
	{
		profileRouter.GET("/:id", func(c *gin.Context) {
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

		profileRouter.PATCH("/:id", r.middleware.RequireUserOwnership("id"), func(c *gin.Context) {
			userID := c.Param("id")

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

		profileImgRouter := profileRouter.Group("/image")
		profileImgRouter.Use(r.middleware.RequireUserOwnership("user_id"))
		{
			profileImgRouter.PATCH("/:user_id", func(c *gin.Context) {
				var oldImagePath string
				err := r.pgClient.QueryRow(c.Request.Context(), `
					SELECT profile_image_url FROM user_profiles
					WHERE user_id = $1`, c.Param("user_id")).Scan(&oldImagePath)
				if err != nil {
					if err == pgx.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
						return
					}
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error1"})
					return

				}

				if oldImagePath != "" {
					if err = utils.RemoveProfileImage(oldImagePath); err != nil {
						utils.LogError(c, err)
						//c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove old image"})
						//return
					}
				}

				imageURL, err := utils.UploadProfileImage(c)
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload image"})
					return
				}

				_, err = r.pgClient.Exec(c.Request.Context(), `
					UPDATE user_profiles 
					SET profile_image_url = $1
					WHERE user_id = $2`, imageURL, c.Param("user_id"))
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error2"})
					return
				}

				c.JSON(http.StatusOK, gin.H{"image_url": imageURL})
			})

			profileImgRouter.DELETE("/:user_id", func(c *gin.Context) {
				var oldImagePath string
				err := r.pgClient.QueryRow(c.Request.Context(), `
					SELECT profile_image_url FROM user_profiles
					WHERE user_id = $1`, c.Param("user_id")).Scan(&oldImagePath)
				if err != nil {
					if err == pgx.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
						return
					}
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				if oldImagePath != "" {
					if err = utils.RemoveProfileImage(oldImagePath); err != nil {
						utils.LogError(c, err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove old image"})
						return
					}

					_, err = r.pgClient.Exec(c.Request.Context(), `
					UPDATE user_profiles 
					SET profile_image_url = ''
					WHERE user_id = $1`, c.Param("user_id"))
					if err != nil {
						utils.LogError(c, err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
						return
					}
				}

				c.JSON(http.StatusOK, gin.H{})
			})

			profileRouter.POST("/:id/follow", func(c *gin.Context) {
				toFollowID, err := strconv.Atoi(c.Param("id"))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
					return
				}

				userThatFollowsID, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
					return
				}
				println(userThatFollowsID)

				if toFollowID == 0 || userThatFollowsID == 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
					return
				}

				if toFollowID == userThatFollowsID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot follow yourself"})
					return
				}

				_, err = r.pgClient.Exec(c.Request.Context(), `
					INSERT INTO follows (profile_id, follower_id) 
					VALUES ($1, $2)`, toFollowID, userThatFollowsID)
				if err != nil {
					if err == pgx.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
						return
					}
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				c.JSON(http.StatusOK, gin.H{})
			})
		}
	}
}
