package routes

import (
	"net/http"

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

			c.SetCookie("AUTH", "", -1, "/", "", false, true)
			c.JSON(http.StatusOK, gin.H{"message": "Account removed successfully"})
		})
	}
}
