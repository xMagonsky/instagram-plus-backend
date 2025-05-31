package routes

import (
	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (r *RoutesManager) RegisterReportsRoutes(router *gin.Engine) {
	reportsRouter := router.Group("/reports")
	reportsRouter.Use(r.middleware.RequireAuth())
	{
		reportsRouter.POST("/user/:username", func(c *gin.Context) {
			var req models.ReportedRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			username := c.Param("username")
			var userID int
			err := r.pgClient.QueryRow(c.Request.Context(), "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}
			reporterID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}
			_, err = r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO reported_users (user_id, reporter_id, reason) VALUES ($1, $2, $3)", userID, reporterID, req.Reason)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})

		reportsRouter.POST("/post/:post_id", func(c *gin.Context) {
			var req models.ReportedRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			postIDStr := c.Param("post_id")
			postID, err := strconv.Atoi(postIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post_id"})
				return
			}
			reporterID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}
			_, err = r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO reported_post (post_id, reporter_id, reason) VALUES ($1, $2, $3)", postID, reporterID, req.Reason)
			if utils.IsForeignKeyViolationPgxError(err, "reported_post_post_id_fkey") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no post with given id"})
				return
			}
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})

		reportsRouter.POST("/comment/:comment_id", func(c *gin.Context) {
			var req models.ReportedRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			commentIDStr := c.Param("comment_id")
			commentID, err := strconv.Atoi(commentIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment_id"})
				return
			}
			reporterID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}
			_, err = r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO reported_comments (comment_id, reporter_id, reason) VALUES ($1, $2, $3)", commentID, reporterID, req.Reason)
			if utils.IsForeignKeyViolationPgxError(err, "reported_comments_comment_id_fkey") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no comment with given id"})
				return
			}
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})

		adminGroup := reportsRouter.Group("")
		adminGroup.Use(r.middleware.RequireAdmin())
		{
			adminGroup.GET("/users", func(c *gin.Context) {
				rows, err := r.pgClient.Query(c.Request.Context(), "SELECT id, user_id, reporter_id, reason FROM reported_users")
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}
				defer rows.Close()
				reports := []models.ReportedUser{}
				for rows.Next() {
					var report models.ReportedUser
					if err := rows.Scan(&report.ID, &report.UserID, &report.ReporterID, &report.Reason); err != nil {
						utils.LogError(c, err)
						continue
					}
					reports = append(reports, report)
				}
				c.JSON(http.StatusOK, reports)
			})

			adminGroup.GET("/posts", func(c *gin.Context) {
				rows, err := r.pgClient.Query(c.Request.Context(), "SELECT id, post_id, reporter_id, reason FROM reported_post")
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}
				defer rows.Close()
				reports := []models.ReportedPost{}
				for rows.Next() {
					var report models.ReportedPost
					if err := rows.Scan(&report.ID, &report.PostID, &report.ReporterID, &report.Reason); err != nil {
						utils.LogError(c, err)
						continue
					}
					reports = append(reports, report)
				}
				c.JSON(http.StatusOK, reports)
			})

			adminGroup.GET("/comments", func(c *gin.Context) {
				rows, err := r.pgClient.Query(c.Request.Context(), "SELECT id, comment_id, reporter_id, reason FROM reported_comments")
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}
				defer rows.Close()
				reports := []models.ReportedComment{}
				for rows.Next() {
					var report models.ReportedComment
					if err := rows.Scan(&report.ID, &report.CommentID, &report.ReporterID, &report.Reason); err != nil {
						utils.LogError(c, err)
						continue
					}
					reports = append(reports, report)
				}
				c.JSON(http.StatusOK, reports)
			})
		}
	}
}
