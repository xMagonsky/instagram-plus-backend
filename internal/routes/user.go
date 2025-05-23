package routes

import (
	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (r *RoutesManager) RegisterUserRoutes(router *gin.Engine) {
	profileRouter := router.Group("/profile")
	profileRouter.Use(r.middleware.RequireAuth())
	{
		userIdProfileRouter := profileRouter.Group("/:user_id")
		{

			userIdProfileRouter.PATCH("", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
				userID := c.Param("user_id")

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

			userIdProfileRouter.PATCH("/image", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
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

			userIdProfileRouter.DELETE("/image", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
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
		}

		usernameProfileRouter := profileRouter.Group("/name/:username")
		{
			usernameProfileRouter.GET("", func(c *gin.Context) {
				username := c.Param("username")
				if username == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
					return
				}

				var userID int
				var user models.Profile
				err := r.pgClient.QueryRow(c.Request.Context(), `
					SELECT u.id, u.username, p.name, p.surname, p.description, p.profile_image_url, p.gender, p.birth, u.creation_timestamp
					FROM users u
					JOIN user_profiles p ON u.id = p.user_id
					WHERE u.username = $1`, username).Scan(
					&userID, &user.Username, &user.Name, &user.Surname, &user.Description, &user.ProfileImageURL, &user.Gender, &user.BirthDate, &user.CreationTimestamp)
				if err != nil {
					if err == pgx.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
					} else {
						utils.LogError(c, err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					}
					return
				}

				// Get followers and following count
				var followersCount, followingCount int
				err = r.pgClient.QueryRow(c.Request.Context(), `
					SELECT COUNT(*) FROM follows WHERE profile_id = $1`, userID).Scan(&followersCount)
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				err = r.pgClient.QueryRow(c.Request.Context(), `
					SELECT COUNT(*) FROM follows WHERE follower_id = $1`, userID).Scan(&followingCount)
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				user.FollowersCount = followersCount
				user.FollowingCount = followingCount

				currentUserID, exists := c.Get("user_id")
				alreadyFollowed := false
				if exists {
					err = r.pgClient.QueryRow(c.Request.Context(), `
						SELECT EXISTS(SELECT 1 FROM follows WHERE profile_id = $1 AND follower_id = $2)
					`, userID, currentUserID).Scan(&alreadyFollowed)
					if err != nil {
						utils.LogError(c, err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
						return
					}
				}

				user.AlreadyFollowed = alreadyFollowed

				c.JSON(http.StatusOK, user)
			})

			usernameProfileRouter.POST("/follow", func(c *gin.Context) {
				toFollowNick := c.Param("username")
				if toFollowNick == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
					return
				}

				userThatFollowsID, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
					return
				}

				var toFollowID int
				err := r.pgClient.QueryRow(c.Request.Context(), `
					SELECT id FROM users WHERE username = $1`, toFollowNick).Scan(&toFollowID)
				if err != nil {
					if err == pgx.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
						return
					}
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

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
					if utils.IsDuplicatePgxError(err) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "you are already following this user"})
						return
					}
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				c.JSON(http.StatusOK, gin.H{})
			})

			usernameProfileRouter.DELETE("/follow", func(c *gin.Context) {
				toUnfollowNick := c.Param("username")
				if toUnfollowNick == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
					return
				}

				userThatUnfollowsID, exists := c.Get("user_id")
				if !exists {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
					return
				}

				var toUnfollowID int
				err := r.pgClient.QueryRow(c.Request.Context(), `
					SELECT id FROM users WHERE username = $1`, toUnfollowNick).Scan(&toUnfollowID)
				if err != nil {
					if err == pgx.ErrNoRows {
						c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
						return
					}
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				if toUnfollowID == 0 || userThatUnfollowsID == 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "user id is required"})
					return
				}

				if toUnfollowID == userThatUnfollowsID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot unfollow yourself"})
					return
				}

				_, err = r.pgClient.Exec(c.Request.Context(), `
					DELETE FROM follows 
					WHERE profile_id = $1 AND follower_id = $2`, toUnfollowID, userThatUnfollowsID)
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}

				c.JSON(http.StatusOK, gin.H{})
			})
		}
	}
}
