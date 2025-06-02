package routes

import (
	"net/http"
	"strconv"

	"instagramplusbackend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (r *RoutesManager) RegisterAccountRoutes(router *gin.Engine) {
	accountRouter := router.Group("/account")
	accountRouter.Use(r.middleware.RequireAuth())
	{
		accountRouter.DELETE("/remove/:user_id", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
			userID := c.Param("user_id")

			var profileImage string
			err := r.pgClient.QueryRow(c.Request.Context(), `SELECT profile_image_url FROM user_profiles WHERE user_id = $1`, userID).Scan(&profileImage)
			if err != nil && err != pgx.ErrNoRows {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			if profileImage != "" {
				_ = utils.RemoveProfileImage(profileImage)
			}

			// Delete user from users table (CASCADE should handle related data)
			_, err = r.pgClient.Exec(c.Request.Context(), `DELETE FROM users WHERE id = $1`, userID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Account removed successfully"})
		})

		// Change password endpoint
		accountRouter.POST("/change-password/:user_id", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
			userID := c.Param("user_id")
			var req struct {
				OldPassword string `json:"old_password" binding:"required"`
				NewPassword string `json:"new_password" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			userIDInt, _ := strconv.Atoi(userID)
			err := r.auth.ChangePassword(c.Request.Context(), userIDInt, req.OldPassword, req.NewPassword)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
		})

		// Change email endpoint
		accountRouter.POST("/change-email/:user_id", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
			userID := c.Param("user_id")
			var req struct {
				Password string `json:"password" binding:"required"`
				NewEmail string `json:"new_email" binding:"required,email"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			userIDInt, _ := strconv.Atoi(userID)
			err := r.auth.ChangeEmail(c.Request.Context(), userIDInt, req.Password, req.NewEmail)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Email changed successfully"})
		})

		// Set premium status endpoint
		accountRouter.POST("/getpremium/:user_id", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
			userID := c.Param("user_id")
			_, err := r.pgClient.Exec(c.Request.Context(), `UPDATE users SET is_premium = TRUE WHERE id = $1`, userID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Premium enabled"})
		})

		// Remove premium status endpoint
		accountRouter.DELETE("/getpremium/:user_id", r.middleware.RequireUserOwnership("user_id"), func(c *gin.Context) {
			userID := c.Param("user_id")
			_, err := r.pgClient.Exec(c.Request.Context(), `UPDATE users SET is_premium = FALSE WHERE id = $1`, userID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Premium disabled"})
		})
	}
}
