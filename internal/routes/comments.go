package routes

import (
	"instagramplusbackend/internal/models"
	"instagramplusbackend/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

func (r *RoutesManager) RegisterCommentsRoutes(router *gin.Engine) {
	commentsRouter := router.Group("/posts/:post_id/comments")
	commentsRouter.Use(r.middleware.RequireAuth())
	{
		commentsRouter.GET("", func(c *gin.Context) {
			postID, err := strconv.Atoi(c.Param("post_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
				return
			}

			rows, err := r.pgClient.Query(c.Request.Context(), `
				SELECT c.id, c.post_id, c.author_id, u.username, up.profile_image_url, c.content, c.creation_timestamp
				FROM comments c
				JOIN users u ON c.author_id = u.id
				JOIN user_profiles up ON up.user_id = u.id
				WHERE c.post_id = $1
				ORDER BY c.creation_timestamp ASC`, postID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			defer rows.Close()

			comments := []models.Comment{}
			for rows.Next() {
				var comment models.Comment
				err := rows.Scan(&comment.ID, &comment.PostID, &comment.AuthorID, &comment.AuthorUsername, &comment.AuthorProfileImageURL, &comment.Content, &comment.CreationTimestamp)
				if err != nil {
					utils.LogError(c, err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
					return
				}
				comments = append(comments, comment)
			}
			c.JSON(http.StatusOK, comments)
		})

		commentsRouter.POST("", func(c *gin.Context) {
			postID, err := strconv.Atoi(c.Param("post_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid post id"})
				return
			}
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				return
			}
			var req models.AddCommentRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			_, err = r.pgClient.Exec(c.Request.Context(),
				"INSERT INTO comments (post_id, author_id, content) VALUES ($1, $2, $3)",
				postID, userID, req.Content)
			if err != nil {
				if utils.IsForeignKeyViolationPgxError(err, "comments_post_id_fkey") {
					c.JSON(http.StatusBadRequest, gin.H{"error": "no post with id " + c.Param("post_id")})
					return
				}
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})

		commentsRouter.GET(":comment_id", func(c *gin.Context) {
			commentID, err := strconv.Atoi(c.Param("comment_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
				return
			}
			row := r.pgClient.QueryRow(c.Request.Context(), `
				SELECT c.id, c.post_id, c.author_id, u.username, up.profile_image_url, c.content, c.creation_timestamp
				FROM comments c
				JOIN users u ON c.author_id = u.id
				JOIN user_profiles up ON up.user_id = u.id
				WHERE c.id = $1`, commentID)
			var comment models.Comment
			err = row.Scan(&comment.ID, &comment.PostID, &comment.AuthorID, &comment.AuthorUsername, &comment.AuthorProfileImageURL, &comment.Content, &comment.CreationTimestamp)
			if err != nil {
				if err == pgx.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "comment not found"})
					return
				}
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, comment)
		})

		commentsRouter.PUT(":comment_id", r.middleware.RequireCommentOwnership("comment_id"), func(c *gin.Context) {
			commentID, err := strconv.Atoi(c.Param("comment_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
				return
			}
			var req models.UpdateCommentRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
				return
			}
			_, err = r.pgClient.Exec(c.Request.Context(),
				"UPDATE comments SET content = $1 WHERE id = $2",
				req.Content, commentID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})

		commentsRouter.DELETE(":comment_id", r.middleware.RequireCommentOwnership("comment_id"), func(c *gin.Context) {
			commentID, err := strconv.Atoi(c.Param("comment_id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid comment id"})
				return
			}
			_, err = r.pgClient.Exec(c.Request.Context(), "DELETE FROM comments WHERE id = $1", commentID)
			if err != nil {
				utils.LogError(c, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		})
	}
}
